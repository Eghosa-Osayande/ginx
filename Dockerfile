#Get the go version we need the application to bootsrap on
FROM golang:1.23.3-alpine

#RUN go get -u github.com/mitranim/gow

#set parent directory
WORKDIR /app

#copy go.mod to image parent dir
COPY go.mod .

#copy go.sum to image parent dir
COPY go.sum .

#download all dependencies
#RUN go mod download

#install all dependencies in the go.mod file
RUN go mod tidy

#copy all files to the parent dir
COPY . .

RUN export PATH=/go/bin

# RUN go build -o gigmile-hermes .

#serve the app
CMD ["go","run","."]
#CMD ["gow", "run", "."]