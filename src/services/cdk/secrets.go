package cdk

import (
	"context"
	"fmt"

	"github.com/treenq/treenq/src/domain"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
)

func (k *Kube) StoreSecret(ctx context.Context, rawConfig string, space, repoID, key, value string) error {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	// Check if namespace exists, create if not
	_, err = clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, createErr := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
			if createErr != nil {
				return fmt.Errorf("failed to create namespace %s: %w", namespace, createErr)
			}
		} else {
			return fmt.Errorf("failed to get namespace %s: %w", namespace, err)
		}
	}

	existingSecret, err := secretClient.Get(ctx, secretObjectName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretObjectName,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					key: []byte(value),
				},
			}
			_, err = secretClient.Create(ctx, newSecret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create secret %s in namespace %s: %w", secretObjectName, namespace, err)
			}
			return nil
		}
		return fmt.Errorf("failed to get secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	if existingSecret.Data == nil {
		existingSecret.Data = make(map[string][]byte)
	}
	existingSecret.Data[key] = []byte(value)

	_, err = secretClient.Update(ctx, existingSecret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	return nil
}

func (k *Kube) GetSecret(ctx context.Context, rawConfig string, space, repoID, key string) (string, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return "", fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	secret, err := secretClient.Get(ctx, secretObjectName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", domain.ErrSecretNotFound
		}
		return "", fmt.Errorf("failed to get secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	value, ok := secret.Data[key]
	if !ok {
		return "", domain.ErrSecretNotFound
	}

	return string(value), nil
}

func (k *Kube) RemoveSecret(ctx context.Context, rawConfig string, space, repoID, key string) error {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	namespace := ns(space, repoID)
	secretObjectName := secretName(repoID, key)
	secretClient := clientset.CoreV1().Secrets(namespace)

	err = secretClient.Delete(ctx, secretObjectName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete secret %s in namespace %s: %w", secretObjectName, namespace, err)
	}

	return nil
}
