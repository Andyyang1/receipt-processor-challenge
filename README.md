can run to make sure dependency are available
go mod tidy

run:
go run main.go

while app is running, in another terminal window run the following command for POST endpoint

curl -X POST http://localhost:8080/receipts/process 
     -H "Content-Type: application/json" 
     -d @path/to/simple-receipt.json


after running post, run the following command for GET endpoint, replacing the {id} with the id generated from POST

curl -X GET http://localhost:8080/receipts/{id}/points


    