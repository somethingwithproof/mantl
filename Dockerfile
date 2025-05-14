FROM alpine:3.19

RUN apk add --no-cache bash build-base curl git libffi-dev openssh openssl-dev py3-pip python3 python3-dev unzip \
	&& git clone https://github.com/mantl/mantl /mantl \
	&& pip3 install --no-cache-dir -r /mantl/requirements.txt \
	&& apk del build-base python3-dev

VOLUME /local
ENV MANTL_CONFIG_DIR /local

VOLUME /root/.ssh

ENV TERRAFORM_VERSION 1.6.3
RUN mkdir -p /tmp/terraform/ && \
    cd /tmp/terraform/ && \
    curl -SLO https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    cd /usr/local/bin/ && \
    unzip /tmp/terraform/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    rm -rf /tmp/terraform/
ENV TERRAFORM_STATE $MANTL_CONFIG_DIR/terraform.tfstate

WORKDIR /mantl
ENTRYPOINT ["/usr/bin/ssh-agent", "-t", "3600", "/bin/sh", "-c"]
