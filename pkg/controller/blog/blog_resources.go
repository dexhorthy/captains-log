package blog

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/dexhorthy/captains-log/pkg/apis/blogging/v1alpha1"
	"io"
	v13 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *ReconcileBlog) buildSiteConfigMap(instance *v1alpha1.Blog, configMapName string) (*v1.ConfigMap, string, error) {
	var tpl bytes.Buffer
	if err := ConfigTOMLTemplate.Execute(&tpl, instance); err != nil {
		return nil, "", err
	}
	configTOML := tpl.String()
	cm := &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      configMapName,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"config.toml": configTOML,
		},
	}
	h := sha1.New()
	if _, err := io.WriteString(h, configTOML); err != nil {
		return nil, "", err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	return cm, hash, nil
}

func (r *ReconcileBlog) buildContentConfigMap(instance *v1alpha1.Blog, configMapName string) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      configMapName,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{},
	}
	return cm
}

func (r *ReconcileBlog) buildService(deploymentName string, instance *v1alpha1.Blog) *v1.Service {
	svcType := instance.Spec.ServiceType
	if svcType == "" {
		svcType = "LoadBalancer"
	}
	svc := &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"deployment": deploymentName,
			},
			Type: svcType,
			Ports: []v1.ServicePort{
				{
					Port: 1313,
				},
			},
		},
	}
	return svc
}

func (r *ReconcileBlog) buildDeployemnt(deploymentName string, instance *v1alpha1.Blog, siteConfigMapName string, contentConfigMapName string, siteConfigHash string) *v13.Deployment {
	siteVolumeName := instance.Name + "-site"
	contentVolumeName := instance.Name + "-content"
	cacheVolumeName := "cachedir"
	var replicas int32 = 2

	deploy := &v13.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      deploymentName,
			Namespace: instance.Namespace,
		},
		Spec: v13.DeploymentSpec{
			Selector: &v12.LabelSelector{
				MatchLabels: map[string]string{"deployment": deploymentName},
			},
			Replicas: &replicas,
			Strategy: v13.DeploymentStrategy{
				RollingUpdate: &v13.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v12.ObjectMeta{
					Labels: map[string]string{
						"deployment":   deploymentName,
						"captains-log": "true",
					},
				},
				Spec: v1.PodSpec{
					Volumes:    r.buildVolumes(siteVolumeName, siteConfigMapName, contentVolumeName, contentConfigMapName, cacheVolumeName),
					Containers: r.buildContainers(instance, siteVolumeName, contentVolumeName, siteConfigHash),
				},
			},
		},
	}
	return deploy
}

func (r *ReconcileBlog) buildContainers(instance *v1alpha1.Blog, siteVolumeName string, contentVolumeName string, siteConfigHash string) []v1.Container {
	containers := []v1.Container{
		{
			Name:  instance.Name,
			Image: "dexhorthy/captains-log-hugo:1",
			Ports: []v1.ContainerPort{
				{
					ContainerPort: 1313,
				},
			},
			Command: []string{
				"/bin/sh", "-c",
				`cp -r /src /site; /hugo server --source=/site --themesDir=/themes --bind=0.0.0.0`,
			},
			Env: []v1.EnvVar{
				{
					Name:  "_SITE_CONFIG_HASH",
					Value: siteConfigHash,
				},
			},
			WorkingDir: "/",
			ReadinessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 1313,
						},
					},
				},
			},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      siteVolumeName,
					MountPath: "/src",
				},
				{
					Name:      contentVolumeName,
					MountPath: "/src/content",
				},
			},
		},
	}
	return containers
}

func (r *ReconcileBlog) buildVolumes(siteVolumeName string, siteConfigMapName string, contentVolumeName string, contentConfigMapName string, cacheVolumeName string) []v1.Volume {
	volumes := []v1.Volume{
		{
			Name: siteVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: siteConfigMapName,
					},
				},
			},
		},
		{
			Name: contentVolumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: contentConfigMapName,
					},
				},
			},
		},
		{
			Name: cacheVolumeName,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/tmp",
				},
			},
		},
	}
	return volumes
}
