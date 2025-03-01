package specs

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func camundaExporterEnv(hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	return exporterEnv("CAMUNDAEXPORTER", "io.camunda.exporter.CamundaExporter", "CONNECT", hostName, username, password)
}

func elasticsearchExporterEnv(hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	els := exporterEnv("ELASTICSEARCH", "io.camunda.zeebe.exporter.ElasticsearchExporter", "AUTHENTICATION", hostName, username, password)
	els[1].Name = "ZEEBE_BROKER_EXPORTERS_ELASTICSEARCH_ARGS_URL"
	return els
}

func exporterEnv(name string, className string, param string, hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  fmt.Sprintf("ZEEBE_BROKER_EXPORTERS_%s_CLASS_NAME", name),
			Value: className,
		},
		{
			Name:  fmt.Sprintf("ZEEBE_BROKER_EXPORTERS_%s_ARGS_%s_URL", name, param),
			Value: hostName,
		},
		{
			Name:  fmt.Sprintf("ZEEBE_BROKER_EXPORTERS_%s_ARGS_%s_USERNAME", name, param),
			Value: username,
		},
		{
			Name:      fmt.Sprintf("ZEEBE_BROKER_EXPORTERS_%s_ARGS_%s_PASSWORD", name, param),
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &password},
		},
	}
}

func zeebeElasticsearch(hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CAMUNDA_ZEEBE_ELASTICSEARCH_URL",
			Value: hostName,
		},
		{
			Name:  "CAMUNDA_ZEEBE_ELASTICSEARCH_USERNAME",
			Value: username,
		},
		{
			Name:      "CAMUNDA_ZEEBE_ELASTICSEARCH_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &password},
		},
	}
}

func camundaDatabaseElasticsearch(hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CAMUNDA_DATABASE_CONNECT_TYPE",
			Value: "elasticsearch",
		},
		{
			Name:  "CAMUNDA_DATABASE_CONNECT_URL",
			Value: hostName,
		},
		{
			Name:  "CAMUNDA_DATABASE_CONNECT_CLUSTERNAME",
			Value: "elasticsearch",
		},
		{
			Name:  "CAMUNDA_DATABASE_CONNECT_USERNAME",
			Value: username,
		},
		{
			Name:      "CAMUNDA_DATABASE_CONNECT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &password},
		},
	}
}

func operateDatabase(hostName, username string, password corev1.SecretKeySelector) []corev1.EnvVar {
	const app = "OPERATE"
	return []corev1.EnvVar{
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_DATABASE", app),
			Value: "elasticsearch",
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ELASTICSEARCH_URL", app),
			Value: hostName,
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ELASTICSEARCH_PREFIX", app),
			Value: "zeebe-record",
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ELASTICSEARCH_CLUSTERNAME", app),
			Value: "elasticsearch",
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ELASTICSEARCH_USERNAME", app),
			Value: username,
		},
		{
			Name:      fmt.Sprintf("CAMUNDA_%s_ELASTICSEARCH_PASSWORD", app),
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &password},
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ZEEBEELASTICSEARCH_URL", app),
			Value: hostName,
		},
		{
			Name:  fmt.Sprintf("CAMUNDA_%s_ZEEBEELASTICSEARCH_USERNAME", app),
			Value: username,
		},
		{
			Name:      fmt.Sprintf("CAMUNDA_%s_ZEEBEELASTICSEARCH_PASSWORD", app),
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &password},
		},
	}
}
