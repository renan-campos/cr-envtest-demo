package main

import (
	"context"
	"flag"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func main() {
	var useTest = flag.Bool("t", false, "run test using controller-runtime testing environment")
	flag.Parse()

	var testEnv *envtest.Environment
	var cfg *rest.Config
	var err error

	if *useTest {
		testEnv = &envtest.Environment{}

		cfg, err = testEnv.Start()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Using w/e k8s cluster is available, typically minikube.
		cfg = config.GetConfigOrDie()
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatal("Error initing k8s client")
	}

	// Will create two config maps, a and b.
	cmA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-a",
			Namespace: "default",
		},
	}
	cmB := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-b",
			Namespace: "default",
		},
	}

	log.Println("Creating ConfigMap A")
	if err := k8sClient.Create(context.TODO(), cmA); err != nil {
		log.Fatal(err)
	}

	log.Println("Creating ConfigMap B")
	if err := k8sClient.Create(context.TODO(), cmB); err != nil {
		log.Fatal(err)
	}

	// set the owner of config map b = a.
	log.Println("Setting ConfigMap B to have ConfigMap A as its controller reference")
	if err := controllerutil.SetControllerReference(cmA, cmB, scheme.Scheme); err != nil {
		log.Fatal(err)
	}
	if err := k8sClient.Update(context.TODO(), cmB); err != nil {
		log.Fatal(err)
	}

	// Delete config map a
	log.Println("Deleting ConfigMap A")
	if err := k8sClient.Delete(context.TODO(), cmA, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
		log.Fatal(err)
	}

	// Wait until the config map has been deleted.
	log.Println("Waiting for ConfigMap A to be deleted...")
	for i := 0; ; i++ {

		err = k8sClient.Get(
			context.TODO(),
			types.NamespacedName{
				Name:      cmA.Name,
				Namespace: cmA.Namespace,
			},
			cmA)

		if errors.IsNotFound(err) {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			if i == 999 {
				// By this point, about ten seconds have gone by, and the resource hasn't been deleted.
				log.Fatal("Timed out waiting for ConfigMap A to be deleted")
			}
			time.Sleep(10 * time.Millisecond)
		}

	}
	log.Println("ConfigMap A has been deleted")

	// Expectation: config map b will also be deleted.
	log.Println("Verifying that ConfigMap B was also deleted...")
	for i := 0; ; i++ {

		err = k8sClient.Get(
			context.TODO(),
			types.NamespacedName{
				Name:      cmB.Name,
				Namespace: cmB.Namespace,
			},
			cmB)

		if errors.IsNotFound(err) {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			if i == 999 {
				// By this point, about ten seconds have gone by, and the resource hasn't been deleted.
				log.Fatal("Timed out waiting for ConfigMap B to be deleted")
			}
			time.Sleep(10 * time.Millisecond)
		}

	}
	log.Println("ConfigMap B has been deleted")
}
