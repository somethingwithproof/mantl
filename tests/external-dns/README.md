# ExternalDNS E2E (RFC2136)

This test provisions a bind9 server inside the cluster and configures ExternalDNS with RFC2136 + TSIG.
It then creates an annotated Service and validates record creation using dig.

Run locally (kind required):
```
kind create cluster --name mantl-dns
kubectl apply -f tests/external-dns/rfc2136/bind9.yaml
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
helm upgrade --install extdns external-dns/external-dns \
  --namespace external-dns --create-namespace \
  -f tests/external-dns/rfc2136/values.yaml
kubectl apply -f tests/external-dns/rfc2136/app-svc.yaml
kubectl run -n dns dig --image=infoblox/dnstools --restart=Never --rm -it --command -- \
  dig @bind9.dns.svc.cluster.local app.test.local +short
```