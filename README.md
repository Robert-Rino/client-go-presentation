
# How to run
```shell
cd {FOLDER}
go mod tidy
go run main.go
```


# Run dummy pod for testing watch_pod
```
cat << EOF | kubectl apply -f - 
apiVersion: apps/v1
kind: Deployment
metadata:
  name: watch-demo
  labels:
    app: swag
spec:
  replicas: 1
  selector:
    matchLabels:
      app: swag
  template:
    metadata:
      labels:
        app: swag
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
EOF
```
