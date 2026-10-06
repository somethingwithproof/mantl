<!-- SPDX-License-Identifier: Apache-2.0 -->
# Example transport prerequisites

These examples require provisioned TLS endpoints. They do not install or claim
successful operation of Tempo, Pushgateway or Flagger webhook TLS servers. Missing
trust ConfigMaps prevent the telemetry pods from starting. Do not bypass
verification to make an unconfigured example run.

For ML and products, configure Tempo's OTLP gRPC listener for TLS on port 4317,
with a certificate covering `tempo.mantl-system`, and create the `telemetry-trust` ConfigMap in
each application namespace with a `ca.crt` key containing the issuing trust roots.
The ML exporter requires HTTPS and uses a secure gRPC channel; products supplies
the standard OpenTelemetry endpoint and certificate settings. Products' FastAPI
instrumentation does not itself install an OTLP exporter. Verify export separately
when adding a collector or automatic instrumentation.

For batch processing, configure Pushgateway's HTTPS listener on port 9091 with a
certificate covering `prometheus-pushgateway.mantl-system`. Create
the `pushgateway-trust` ConfigMap in `batch-processing` with the trusted `ca.crt`. The application
rejects non-HTTPS endpoints and urllib's verified HTTPS connection uses the mounted
`SSL_CERT_FILE`. Keep certificate and hostname validation enabled. Configure server
authentication and network access for the deployment's threat model; HTTPS alone
does not authorize a metrics writer.

Before enabling a Canary, replace `loadtester.example.invalid` and
`notifier.example.invalid` in `spec.analysis.webhooks` with separately provisioned
HTTPS services implementing the Flagger load-test and notification contracts. The
reserved `.invalid` defaults deliberately prevent accidental use. Their certificates
must cover the chosen hostnames and chain to trust roots available to the Flagger
controller. Configure that trust in the controller deployment; these application
manifests do not configure it. Run webhook contract and invalid-certificate tests in
a disposable environment before enabling automated rollout. No live acceptance was
performed for these examples.

The stock Flagger loadtester serves HTTP; replacing its URL scheme does not add TLS.
Provide a reviewed TLS frontend colocated with that backend (loopback upstream),
or a HTTPS implementation of the webhook. A TLS proxy forwarding cleartext across
the cluster does not meet this contract. Restrict loadtester access: its command
webhook can launch load-generation commands. The example load-test target remains
HTTP and is appropriate only for synthetic, credential-free test traffic; sensitive
application traffic requires its own TLS configuration.
