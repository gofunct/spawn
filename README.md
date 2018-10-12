# Spawn

Easily test GRPC services locally using docker-compose!

## Building the provided GRPC service

<details><summary>show</summary>
<p>

```
docker build -t spawn:serve .
```

```
docker build -t spawn:client -f Dockerfile.client .
```

</p>
</details>

## Running the GRPC services individually

<details><summary>show</summary>
<p>

```
docker run -d -p 3101:3101 spawn:serve
```

*Cannot run the client, as it depends on a network link to the server*
*Use docker-compos*

</p>
</details>

## 