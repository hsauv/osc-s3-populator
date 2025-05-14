package controllers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storagev1alpha1 "github.com/hsauv/osc-s3-populator/api/v1alpha1"
)

type OutscaleS3PopulatorReconciler struct {
	client.Client
}

//+kubebuilder:rbac:groups=storage.populator.io,resources=outscales3populators,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get

func (r *OutscaleS3PopulatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var populator storagev1alpha1.OutscaleS3Populator
	if err := r.Get(ctx, req.NamespacedName, &populator); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Objet supprimé")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 1. Lire les secrets AK / SK
	accessKey, err := r.getSecretValue(ctx, populator.Namespace, populator.Spec.AccessKeySecretRef)
	if err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur lecture accessKey: %v", err))
	}
	secretKey, err := r.getSecretValue(ctx, populator.Namespace, populator.Spec.SecretKeySecretRef)
	if err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur lecture secretKey: %v", err))
	}

	// 2. Init session AWS S3 compatible
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(populator.Spec.Region),
		Endpoint:         aws.String(populator.Spec.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur session S3: %v", err))
	}

	s3Client := s3.New(sess)

	// 3. Télécharger l’objet
	output, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(populator.Spec.Bucket),
		Key:    aws.String(populator.Spec.Object),
	})
	if err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur téléchargement S3: %v", err))
	}
	defer output.Body.Close()

	// 4. Écriture locale (dans un chemin fixe, ex. /mnt/data)
	destDir := "/mnt/data"
	destFile := filepath.Join(destDir, filepath.Base(populator.Spec.Object))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur création dossier: %v", err))
	}

	file, err := os.Create(destFile)
	if err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur création fichier: %v", err))
	}
	defer file.Close()

	if _, err := io.Copy(file, output.Body); err != nil {
		return r.updateStatus(ctx, &populator, "Failed", fmt.Sprintf("Erreur copie fichier: %v", err))
	}

	// 5. Succès
	return r.updateStatus(ctx, &populator, "Succeeded", "Données téléchargées avec succès")
}

func (r *OutscaleS3PopulatorReconciler) getSecretValue(ctx context.Context, namespace string, selector corev1.SecretKeySelector) (string, error) {
	var secret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{
		Name:      selector.Name,
		Namespace: namespace,
	}, &secret); err != nil {
		return "", err
	}
	value, ok := secret.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("clé '%s' absente dans le secret '%s'", selector.Key, selector.Name)
	}
	return string(value), nil
}

func (r *OutscaleS3PopulatorReconciler) updateStatus(ctx context.Context, pop *storagev1alpha1.OutscaleS3Populator, phase, message string) (ctrl.Result, error) {
	pop.Status.Phase = phase
	pop.Status.Message = message
	if err := r.Status().Update(ctx, pop); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *OutscaleS3PopulatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1alpha1.OutscaleS3Populator{}).
		Complete(r)
}