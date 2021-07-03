how to use :

LD_LIBRARY_PATH=lib dapr run --app-id service-snow  --app-protocol grpc  --app-port 50001  --dapr-grpc-port 3501 --log-level debug --components-path ./config go run main.go