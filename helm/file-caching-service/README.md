# file-caching-service

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)

A Helm chart for file caching service with R2 storage and Redis cache

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+

## Installing the Chart

```bash
helm install my-release ./helm/file-caching-service \
  --set redis.addr=your-redis-host:6379 \
  --set secrets.redisPassword=your-redis-password \
  --set r2.accountId=your-cloudflare-account-id \
  --set r2.bucketName=your-bucket-name \
  --set secrets.r2AccessKeyId=your-access-key \
  --set secrets.r2SecretAccessKey=your-secret-key
```

## Uninstalling the Chart

```bash
helm uninstall my-release
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| replicaCount | int | `2` | Number of replicas |
| image.repository | string | `"file-caching-service"` | Image repository |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.tag | string | `""` | Image tag (defaults to chart appVersion) |
| service.type | string | `"ClusterIP"` | Kubernetes service type |
| service.port | int | `80` | Service port |
| ingress.enabled | bool | `false` | Enable ingress |
| config.port | string | `"8080"` | Application port |
| config.cacheTTL | string | `"5m"` | Cache TTL duration |
| redis.addr | string | `""` | Redis connection address |
| r2.accountId | string | `""` | Cloudflare account ID |
| r2.bucketName | string | `""` | R2 bucket name |
| secrets.redisPassword | string | `""` | Redis password |
| secrets.r2AccessKeyId | string | `""` | R2 access key ID |
| secrets.r2SecretAccessKey | string | `""` | R2 secret access key |

