prepare: swarm-init install-faas-cli

swarm-init:
	docker swarm init --advertise-addr eth0

swarm-leave:
	docker swarm leave --force

docker-login:
	docker login

install-faas-cli:
	curl -sL cli.openfaas.com | sudo sh

deploy-faas:
	(cd ./vendor/github.com/openfaas/faas/; ./deploy_stack.sh)

run-monitor:
	docker service create -d \
		--name=func_grafana \
		--publish=3000:3000 \
		--network=func_functions \
		stefanprodan/faas-grafana:4.6.3

run: deploy-faas run-monitor

down:
	docker swarm leave --force
