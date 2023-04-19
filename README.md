# default-backend

```bash
docker pull ysicing/defaultbackend:0.4.0
# 替换ingress default backend
```

## 使用

```yaml
# kubectl get cm/nginx-ingress-nginx-controller  -n kube-system -o yaml
apiVersion: v1
data:
  access-log-path: /var/log/nginx/nginx_access.log
  compute-full-forwarded-for: "true"
  custom-http-errors: 404,503
  error-log-path: /var/log/nginx/nginx_error.log
  forwarded-for-header: X-Forwarded-For
  generate-request-id: "true"
  keep-alive-requests: "10000"
  log-format-upstream: $remote_addr - $remote_user [$time_iso8601] $msec "$request"
    $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time
    [$proxy_upstream_name] [$proxy_alternative_upstream_name] [$upstream_addr] [$upstream_response_length]
    [$upstream_response_time] [$upstream_status] $req_id
  max-worker-connections: "65536"
  proxy-body-size: 50m
  upstream-keepalive-connections: "200"
  use-forwarded-headers: "true"
kind: ConfigMap
```

## v1版本参考

[defaultbackend-v1](https://github.com/ysicing/dockerfiles/tree/v1.0/defaultbackend)
