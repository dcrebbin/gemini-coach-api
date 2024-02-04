# Gemini Coach API

_Frontend:_ https://github.com/dcrebbin/gemini-coach

This API is for Gemini Coach made with [Go Fiber](https://docs.gofiber.io/)

### Text Generation

- [Vertex (Palm, Gemini Pro etc)](https://console.cloud.google.com/vertex-ai/generative)

### Text to Speech

- [Vertex](https://console.cloud.google.com/vertex-ai/generative)

### Speech to Text

- [Vertex](https://console.cloud.google.com/vertex-ai/generative)

## Setup

1. [Install GO](https://go.dev/doc/install)

1. [Install gcloud CLI](https://cloud.google.com/sdk/docs/install)

1. Create a gcloud project and enable a bunch of things, etc etc

1. Go get

1. Create a .env using the env.example file

## Swagger

_Not fully implemented_

http://127.0.0.1:8080/swagger/index.html

## Authentication

This allows you to deploy to gcp

> gcloud auth login

Need to use auth quickly to use vertex ai?

> gcloud auth print-access-token

## Deploy

(make sure to be within the root directory ./)

> gcloud run deploy --source .
