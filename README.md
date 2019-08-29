# Microservices-shop
Microservice based shopping website written in Python and Go, deployed on Kubernetes in AWS. Personal project to learn Go, DevOps tech and CI/CD.

## Backend

![test](/img/Backend.png)

The application is microservices based and consists of the following services:

| **Service** | **Language** | **Function**                                      |
| ----------- | ------------ | ------------------------------------------------- |
| Frontend    | Go           | HTTP server to serve frontend to users            |
| Carts       | Go           | Handles user shopping carts, stores them in Redis |
| Products    | Go           | Stores all products using Postgres                |
| Checkout    | Go           | Orchestrates product checkout                     |
| Email       | Python       | Sends a fake email to the user                    |
| Shipping    | Python       | Generates a fake shipping ID                      |
| Payment     | Python       | Generates a fake transaction ID                   |

**Tools/frameworks used:**

- [Gin]( https://github.com/gin-gonic/gin) for creating REST APIs
- [Logrus](https://github.com/sirupsen/logrus) for structured logging
- [Mux](https://github.com/gorilla/mux) for the frontend server
- [Starlette](https://www.starlette.io/) as async Python webserver
- Redis as the cart store
- PostgreSQL as the products database



## Infrastructure

The application is container based and deployed in a Kubernetes cluster on AWS. The entire infrastructure is based on Infrastructure as Code principles, using Terraform, Kubernetes, and Ansible.



**Technology and tools used:**

- Kubernetes
- kubeadm
- Helm
- Istio service mesh
- Docker
- Terraform
- Ansible
- AWS



Using Terraform, the following AWS infrastructure is deployed:

![](/img/Infrastructure.png)

This is not a best practice infrastructure due to the usage of an AWS starter account, by which I was limited to only certain AWS resources.

Using Ansible, a Kubernetes cluster with Istio service mesh is deployed. The Kubernetes cluster is formed using the `kubeadm` tool.



## CI / CD

This project uses a CI/CD pipeline using the free tier of CircleCI. The pipeline is described in the `.circleci` folder, and based on the following steps:

1. **Checkout**: using a CircleCI webhook, whenever a commit is pushed to the `master` branch the build pipeline triggers;
2. **Run unit tests**: the Go unit tests are ran to check if the services still work;
3. **Build Docker images**: after successful unit tests, a Docker container is built for each service using the `latest` tag;
4. **Push Docker images**: these images are then pushed to my personal Dockerhub;
5. **Trigger Kubernetes Rolling Update**: Ideally, the next step would be to trigger an update of all containers in Kubernetes. Unfortunately, the free version of CircleCI does not support this.



















