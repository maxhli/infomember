package main

import (
	"log"
	"net/http"
	_ "net/url"
	"os"

	_ "github.com/satori/go.uuid"

	"database/sql"

	_ "github.com/lib/pq"

	//"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/gin-gonic/gin"
	"fmt"

	"context"
	"time"
	_ "io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/credentials"

	_ "github.com/jinzhu/gorm"
	_ "github.com/gin-gonic/gin"
	_ "github.com/aws/aws-sdk-go/private/protocol"
	"strings"
	_ "strconv"
	"strconv"
	"github.com/satori/go.uuid"
	"errors"
)


type Member struct {
	ID   int
	ChineseName  string
	EnglishName string
	Email string
	CellPhone string
	Street string
	City string
	State string
	Zip string
	ShortPixName string
	PictureURL string
}


func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func uploadAFile(c *gin.Context) (string, string, error) {
		// single file
		file, _ := c.FormFile("file")
		log.Println("The file name is: ", file.Filename)

		var bucket, key string
	bucket = "ithreeman"
	bucketPrefix := "https://s3.us-east-2.amazonaws.com/ithreeman/"
		var timeout time.Duration

		timeout = 60 * time.Minute

		AWS_ACCESS_KEY_ID :=
			os.Getenv("AWS-ACCESS-KEY-ID")
		AWS_SECRET_ACCESS_KEY :=
			os.Getenv("AWS-SECRET-ACCESS-KEY")

		token := ""
		creds := credentials.NewStaticCredentials(
			AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, token)

		_, errCred := creds.Get()
		if errCred != nil {
			log.Fatal(errCred)
		}

	session := session.Must(session.NewSession(
			&aws.Config{
				Region:      aws.String(endpoints.UsEast2RegionID),
				Credentials: creds,
			}))
		// Create a new instance of the service's client with a Session.
		// Optional aws.Config values can also be provided as variadic arguments
		// to the New function. This option allows you to provide service
		// specific configuration.
		svc := s3.New(session)

		// Create a context with a timeout that will abort the upload if it takes
		// more than the passed in timeout.
		ctx := context.Background()
		var cancelFn func()
		if timeout > 0 {
			ctx, cancelFn = context.WithTimeout(ctx, timeout)
		}
		// Ensure the context is canceled to prevent leaking.
		// See context package for more information, https://golang.org/pkg/context/
		defer cancelFn()

		//f, errOpen  := os.Open(file.Filename)
		f, errOpen := file.Open()
		if errOpen != nil {
			log.Fatalf("failed to open file %q, %v",
				file.Filename, errOpen)
			return "", "", errors.New("Failed to open file " +
				file.Filename + errOpen.Error())

		}

		key = file.Filename
		t := time.Now()
		var ret string = t.Format(
			"2006-01-02T15:04:05.999999-07:00")
		var year string = ret[0:4]
		var month string = ret[5:7]
		var day string = ret[8:10]

	var keyString string = year + "/" + month +
			"/" + day + "/" + uuid.NewV4().String() + "----" + key


	fmt.Println("It is: ", keyString)


	contentType := ""
	fileExtension := ""

	i := strings.LastIndex(key, ".")
	if i == -1 {
		contentType = "NotAcceptable"
	} else {
		fileExtension = key[i:]
	}

	if fileExtension == ".gif" {
		contentType = "image/gif"
	} else if fileExtension == ".jpeg" {
	contentType = "image/jpeg"
	} else if fileExtension == ".tif" {
		contentType = "image/tiff"
	} else if fileExtension == ".tiff" {
	contentType = "image/tiff"
	} else if fileExtension == ".jpg" {
		contentType = "image/jpg"
	} else if fileExtension == ".png" {
	contentType = "image/png"
	} else if fileExtension == ".bm" {
	    contentType = "image/bmp"
	} else if  fileExtension == ".bmp" {
	contentType = "image/bmp"
	} else {
	    contentType = "NotAcceptable"
	}


	if contentType == "NotAcceptable" {
		return "", "", errors.New("The file type is not acceptable.")

	}

		// Uploads the object to S3. The Context will interrupt the request if the
		// timeout expires.
		_, err := svc.PutObjectWithContext(
			ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(keyString),
			ContentType: aws.String(contentType),
			Body:   f,
		})
		if err != nil {
			aerr, ok := err.(awserr.Error);
			if ok && aerr.Code() ==
				request.CanceledErrorCode {
				// If the SDK can determine the request or retry delay was canceled
				// by a context the CanceledErrorCode error code will be returned.
				log.Fatalf("upload canceled due to timeout, %v\n", err)
				return "", "", errors.New("Upload cancelled due to timeout, " +
					err.Error())
			} else {
				log.Fatalf("failed to upload object, %v\n", err)
				return "", "", errors.New("failed to upload object, " +
					err.Error())
			}
		}
		log.Printf("successfully uploaded file to %s %s\n",
			bucket, keyString)

		return key, bucketPrefix + keyString, nil
		}

func main() {

	var DATABASE_URL = os.Getenv("DATABASE_URL")

	db, errDB := sql.Open("postgres", DATABASE_URL)
	defer db.Close()

	if errDB != nil {
		log.Fatalf("Error connecting to the DB")
	} else {
		log.Println("Connection is successful!")
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")


	router.GET("/members/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "members.create.tmpl.html", nil)
	})

	router.GET("/members/select/:id", func(c *gin.Context) {

		id := c.Param("id")

		rows, err := db.Query("SELECT ID, ChineseName, EnglishName, " +
			" Email, CellPhone, Street, City, State, Zip, " +
				"PictureURL FROM members where ID = $1", id)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		member := new(Member)

		for rows.Next() {
			err := rows.Scan(&member.ID, &member.ChineseName,
				&member.EnglishName, &member.Email,
					&member.CellPhone, &member.Street,
						&member.City, &member.State,
							&member.Zip, &member.PictureURL)
			if err != nil {
				log.Fatal(err)
			}
		}


		c.HTML(http.StatusOK, "members.select.tmpl.html", member)
	})

	router.POST("/members/create", func(c *gin.Context) {
		EnglishName := c.PostForm("EnglishName")
		ChineseName := c.PostForm("ChineseName")
		Email := c.PostForm("Email")
		CellPhone := c.PostForm("CellPhone")
		Street := c.PostForm("Street")
		City := c.PostForm("City")
		State := c.PostForm("State")
		Zip := c.PostForm("Zip")

		//calling uploadAFile to upload it.
		shortPixName, returnedFile, err := uploadAFile(c)

		if err != nil {
			log.Println("Upload an image file encounters a problem.")
			c.HTML(http.StatusOK, "members.create_error.tmpl.html", err)
		}

		fmt.Println("returned file name is : ", returnedFile)


		_, errInsert := db.
		Exec("INSERT INTO members(ChineseName, EnglishName, " +
			"Email, CellPhone, Street, City, State, Zip, " +
				"ShortPixName, PictureURL) VALUES " +
					"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
			ChineseName, EnglishName, Email, CellPhone,
				Street, City, State, Zip, shortPixName, returnedFile)

		if errInsert != nil {
			log.Println("DB Insertion is in error.")
			c.HTML(http.StatusOK,
				"members.create_error.tmpl.html", errInsert)
		} else {
			log.Println("DB Insertion successful.")
			rows, err := db.Query("SELECT ID, ChineseName, " +
				"EnglishName, Email, CellPhone, Street, City, State, " +
					"Zip, PictureURL FROM members order by ID DESC")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()


			c.HTML(http.StatusOK, "members.create_ok.tmpl.html", nil)
		}
	})

	router.GET("/members/update/:id", func(c *gin.Context) {
		id := c.Param("id")

		rows, err := db.Query("SELECT ID, ChineseName, EnglishName, " +
			"Email, CellPhone, Street, City, State, zip, " +
			"ShortPixName, PictureURL FROM members where ID = $1", id)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		member := new(Member)

		for rows.Next() {
			err := rows.Scan(&member.ID, &member.ChineseName, &member.EnglishName,
				&member.Email, &member.CellPhone,
				&member.Street, &member.City,
				&member.State, &member.Zip,
				&member.ShortPixName, &member.PictureURL)
			if err != nil {
				log.Fatal(err)
			}
		}
		c.HTML(http.StatusOK, "members.update.tmpl.html", member)

	})

	router.POST("/members/update/:id", func(c *gin.Context) {
		ID := c.Param("id")

		IDNumber, err1 := strconv.Atoi(ID)
		checkErr(err1)

		EnglishName := c.PostForm("EnglishName")
		ChineseName := c.PostForm("ChineseName")
		Email := c.PostForm("Email")
		CellPhone := c.PostForm("CellPhone")
		Street := c.PostForm("Street")
		City := c.PostForm("City")
		State := c.PostForm("State")
		Zip := c.PostForm("Zip")


		DistanceFromChurch := c.PostForm("DistanceFromChurch")

		// Update
		stmt, err := db.Prepare(
			"update members set EnglishName = $1, ChineseName = $2, " +
				"Email = $3, CellPhone = $4, " +
		        "Street = $5, City = $6, " +
		        "State = $7, Zip = $8, " +
		        "DistanceFromChurch = $9 where ID=$10")
		checkErr(err)
		fmt.Println("update statement is: ", stmt)

		val, err := strconv.ParseFloat(DistanceFromChurch, 32)

		fmt.Println("EnglishName, ChineseName, val, ID are: ", EnglishName, ChineseName, val, ID)

		res, err2 := stmt.Exec(EnglishName, ChineseName, Email, CellPhone,
			Street, City, State, Zip, val, IDNumber)

		checkErr(err2)
		defer stmt.Close()

		rowsAffected, err3 := res.RowsAffected()
		checkErr(err3)
		fmt.Println("rowsAffected is: ", rowsAffected)



		c.HTML(http.StatusOK, "members.update_post.tmpl.html", ID)

	})

	router.GET("/members/delete/:id", func(c *gin.Context) {
		ID := c.Param("id")

		member := new(Member)
		idHolder, err1 := strconv.Atoi(ID)
		member.ID = idHolder

		if err1 != nil {
			panic("ID is not a big integer. Terribly wrong")
		}

		c.HTML(http.StatusOK, "members.delete.tmpl.html", member)

	})

	router.POST("/members/delete/:id", func(c *gin.Context) {
		ID := c.Param("id")

		// Update
		stmt, err := db.Prepare(
			"delete from Members where ID=$1")
		checkErr(err)
		fmt.Println("statement is: ", stmt)
		fmt.Println("ID is: ", ID)

		res, err2 := stmt.Exec(ID)
		checkErr(err2)
		defer stmt.Close()

		rowsAffected, err3 := res.RowsAffected()
		checkErr(err3)
		fmt.Println("rowsAffected is: ", rowsAffected)

		c.HTML(http.StatusOK, "members.delete_post.tmpl.html", ID)

	})

	router.GET("/", func(c *gin.Context) {
		rows, err := db.Query("SELECT ID, ChineseName, " +
			"EnglishName, Email, CellPhone, ShortPixName, " +
				" PictureURL FROM Members order by ID DESC")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		members := make([]*Member, 0)
		for rows.Next() {
			member := new(Member)
			err := rows.Scan(&member.ID, &member.ChineseName, &member.EnglishName,
				&member.Email, &member.CellPhone,
					&member.ShortPixName, &member.PictureURL)
			if err != nil {
				log.Fatal(err)
			}
			members = append(members, member)
		}
		c.HTML(http.StatusOK, "index.tmpl.html", members)

	})

	router.Run(":" + port)
}
