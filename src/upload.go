package src

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func ConnectDB() {
	connectionString := "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	db = client.Database("your_database_name")
}

func UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	bucket, err := gridfs.NewBucket(db)
	if err != nil {
		http.Error(w, fmt.Sprintf("не удалось создать GridFS bucket: %v", err), http.StatusInternalServerError)
		return
	}

	uploadStream, err := bucket.OpenUploadStream("image.jpg", options.GridFSUpload().SetMetadata(bson.D{{"type", "image"}}))
	if err != nil {
		http.Error(w, fmt.Sprintf("не удалось открыть поток для загрузки: %v", err), http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("не удалось прочитать файл: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = uploadStream.Write(fileBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf("не удалось записать данные в GridFS: %v", err), http.StatusInternalServerError)
		return
	}

	fileID := uploadStream.FileID.(primitive.ObjectID)
	fmt.Fprintf(w, `{"image_id": "%s"}`, fileID.Hex())
}
