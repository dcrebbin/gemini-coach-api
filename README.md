# Gemini Coach API

_Frontend:_ https://github.com/dcrebbin/gemini-coach

This API is for Gemini Coach made with [Go Fiber](https://docs.gofiber.io/)

### Text Generation

- [Vertex (Palm2, Gemini Pro etc)](https://console.cloud.google.com/vertex-ai/generative)

### Text to Speech

- [Vertex](https://console.cloud.google.com/vertex-ai/generative)

### Speech to Text

- [Vertex](https://console.cloud.google.com/vertex-ai/generative)

## Setup

1. [Install Go](https://go.dev/doc/install)

2. Create a gcloud project and enable VertexAi

3. Create a .env using the env.example file

4. Navigate to `https://console.cloud.google.com/apis/credentials?authuser=1&project=` and create a service account and give it access to the VertexAi role via Google IAM

5. Download those credentials and store in `/authentication/*your credentials*.json`

6. go get

7. go run main.go

## Deploy to production

1. head to https://console.cloud.google.com/run?hl=en&project=*your project*

2. Create service > Continuously deploy from a repository (Github) > Set up with cloud build > Install & enable G Cloud Build for your repo > build type "Dockerfile" > Save

3. Click into your service > Edit & Deploy New Revision > Variables & Secrets > Enter your variables & Secrets

4. Deploy


_Secret Manager Not fully implemented_


## Swagger

_Not implemented_

http://127.0.0.1:8080/swagger/index.html
