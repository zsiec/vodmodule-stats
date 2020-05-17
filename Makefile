run:
	go run ./cmd/scrape/main.go

image:
	docker build -t zsiec/vodmodule-stats:latest -f Dockerfile .

push:
	docker push zsiec/vodmodule-stats:latest

release-dev:
	aws eks --region us-east-1 update-kubeconfig --name vpt-eks-dev && \
	kubectl apply -f ./deploy/k8s/dev

deploy-dev: image push release-dev

destroy-dev:
	aws eks --region us-east-1 update-kubeconfig --name vpt-eks-dev && \
	kubectl delete -f ./deploy/k8s/dev
