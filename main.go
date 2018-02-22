package main

import (
	"log"
	"net/http"
	_ "net/url"
	"os"

	_ "github.com/satori/go.uuid"

	"github.com/gorilla/sessions"

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
	"golang.org/x/crypto/bcrypt"
	_ "net/smtp"
	_ "github.com/qor/roles"
)

type User struct {
	ID int
	Username string
	FirstName string
	LastName string
	Email string
	Password string
}

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
	Disabled bool
	UsernameAsOwner string
}
var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key = []byte("THE-KEY-has-to-be-32-bytes-long!")
	store = sessions.NewCookieStore(key)
)
func isAuth(c *gin.Context) (bool, string) {
	session1, _ := store.Get(
		c.Request, "infomember-cookie")
	auth, ok := session1.Values["authenticated"].(bool)
	// Check if user is authenticated
	if !ok || !auth {
		return false, ""
	} else {
		return true, session1.Values["username"].(string)
	}
}
func checkAuth(c *gin.Context) {
	session1, _ := store.Get(
		c.Request, "infomember-cookie")
	// Check if user is authenticated
	if auth, ok := session1.Values["authenticated"].
	(bool); !ok || !auth {
		http.Error(c.Writer, "Forbidden",
			http.StatusForbidden)
		return
	}
}

func secret(w http.ResponseWriter, r *http.Request) {
	session1, _ := store.Get(
		r, "infomember-cookie")

	// Check if user is authenticated
	if auth, ok := session1.Values["authenticated"].
		(bool); !ok || !auth {
		http.Error(w, "Forbidden",
			http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "In a secret area.")
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

	sessionAWS := session.Must(session.NewSession(
			&aws.Config{
				Region:      aws.String(endpoints.UsEast2RegionID),
				Credentials: creds,
			}))
		// Create a new instance of the service's client with a Session.
		// Optional aws.Config values can also be provided as variadic arguments
		// to the New function. This option allows you to provide service
		// specific configuration.
		svc := s3.New(sessionAWS)

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

func getPwd() []byte {
	// Prompt the user to enter a password
	fmt.Println("Enter a password:")
	// We will use this to store the users input
	var pwd string
	// Read the users input
	_, err := fmt.Scan(&pwd)
	if err != nil {
		log.Println(err)
	}
	// Return the users input as a byte slice which will save us
	// from having to do this conversion later on
	return []byte(pwd)
}

func hashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func doesPermissionCodeExist (db *sql.DB, email string,
	permissioncode string) bool {
	userRows, errSelect := db.Query(
		"select count(*) from emailpermissioncodes " +
			"where email = $1 and permissioncode = $2",
			email, permissioncode)

	if errSelect != nil {
		log.Println("Selection from emailpermissioncodes " +
			" table is NOT successful.")
		return false
	} else {
		var ret int = 0
		for userRows.Next() {
			err := userRows.Scan(&ret)
			if err != nil {
				log.Fatal(err)
				return false
			}
		}
		if ret == 1 {
			return true
		} else {
			return false
		}

	}
}

func doesUsernameExist (db *sql.DB, username string) bool {
	userRows, errSelect := db.Query(
		"select username from users " +
			"where username = $1", username)

	if errSelect != nil {
		log.Println("Selection password hash from users table is NOT successful.")
		return false
	} else {
		user1 := new(User)
		defer userRows.Close()
		hasData := false
		for userRows.Next() {
			hasData = true
			err := userRows.Scan(&user1.Username)
			if err != nil {
				log.Fatal(err)
				return false
			}
		}
		if hasData {
			return true
		} else {
			return false
		}

	}
}

func main() {

	var DATABASE_URL = os.Getenv("DATABASE_URL")

	db, errDB := sql.Open("postgres", DATABASE_URL)
	defer db.Close()

	if errDB != nil {
		log.Fatalf("Error connecting to the DB.")
	} else {
		log.Println("Connection is successful!")
	}

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")


	router.GET("/members/create", func(c *gin.Context) {
		// checkAuth(c)
		authenticated, username := isAuth(c)
		if !authenticated {
			c.String(http.StatusOK, "You need to login at first." + username)
		}else {
			c.HTML(http.StatusOK, "members.create.tmpl.html", username)
		}
	})

	router.GET("/members/select/:id", func(c *gin.Context) {
		checkAuth(c)

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
		checkAuth(c)

		session1, _ := store.Get(
			c.Request, "infomember-cookie")
		username, _ := session1.Values["username"].(string)

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
				"ShortPixName, PictureURL, UsernameAsOwner) VALUES " +
					"($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			ChineseName, EnglishName, Email, CellPhone,
				Street, City, State, Zip,
					shortPixName, returnedFile, username)

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
		checkAuth(c)
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
		checkAuth(c)
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

	router.GET("/accounts/create",
		func(c *gin.Context) {
			c.HTML(http.StatusOK,
				"accounts.create.tmpl.html", nil)
	})

	router.POST("/accounts/create", func(c *gin.Context) {
		username := c.PostForm("username")
		firstName := c.PostForm("first_name")
		lastName := c.PostForm("last_name")
		email := c.PostForm("email")
		password1 := c.PostForm("password1")
		password2 := c.PostForm("password2")
		permissioncode := c.PostForm("permissioncode")

		log.Println("username is: ", username, "first_name is: ", firstName,
			"lastName is: ", lastName, "password1 is: ", password1,
			"password2 is: ", password2,
				"permissioncode is: ", permissioncode)


		if !doesPermissionCodeExist(db, email, permissioncode) {
			emsg := "Permission code is wrong. Please contact " +
				" the system administrator."
			log.Println(emsg)
			c.HTML(http.StatusOK,
				"accounts.create_error.tmpl.html", emsg)
		} else if doesUsernameExist(db, username) {
			emsg := "Username already exists. Please choose another one."
			log.Println(emsg)
			c.HTML(http.StatusOK,
				"accounts.create_error.tmpl.html", emsg)
		} else {
			_, errInsert := db.
				Exec("INSERT INTO users(username, firstname, lastname, email,"+
				"password) VALUES "+
				"($1, $2, $3, $4, $5)",
				username, firstName, lastName, email,
				hashAndSalt([]byte(password1)))

			if errInsert != nil {
				log.Println("DB Insertion is in error.")
				c.HTML(http.StatusOK,
					"accounts.create_error.tmpl.html", errInsert)
			} else {
				log.Println("DB Insertion successful.")
				c.HTML(http.StatusOK, "accounts.create_ok.tmpl.html", nil)
			}
	    }
	})

	router.GET("/accounts/logout", func(c *gin.Context) {
		c.HTML(http.StatusOK, "accounts.logout.tmpl.html", nil)
	})

	router.POST("/accounts/logout", func(c *gin.Context) {
		log.Println("log out!!!")
		session1, _ := store.Get(
			c.Request, "infomember-cookie")

		// Revoke users authentication
		session1.Values["authenticated"] = false
		session1.Save(c.Request, c.Writer)
		c.Redirect(http.StatusFound, "/")
	})

	router.GET("/accounts/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "accounts.login.tmpl.html", nil)
	})

	router.POST("/accounts/login", func(c *gin.Context) {
		username := c.PostForm("username")
		enteredPWD := c.PostForm("password")

		log.Println("username is: ", username,
			"password is: ", enteredPWD)

		userRows, errSelect := db.Query(
			"select ID, username, password from users " +
			"where username = $1", username)

		if errSelect != nil {
			log.Println("Selection password hash from users table is NOT successful.")
			c.HTML(http.StatusOK,
				"accounts.login_error.tmpl.html", errSelect)
		} else {

			user1 := new(User)

			defer userRows.Close()

			hasData := false
			for userRows.Next() {
				hasData = true
				err := userRows.Scan(&user1.ID, &user1.Username, &user1.Password)
				if err != nil {
					log.Fatal(err)
				}
			}

			if hasData {
				log.Println("Got data from DB table.")
				if comparePasswords(user1.Password, []byte(enteredPWD)) {
					log.Println("Login sucessful!")
					infoMsg := "Login sucessful!"
					session1, _ := store.Get(
						c.Request, "infomember-cookie")
					session1.Values["authenticated"] = true
					session1.Values["username"] = user1.Username
					session1.Save(c.Request, c.Writer)
					c.HTML(http.StatusOK,
						"accounts.login_ok.tmpl.html", infoMsg)

				} else {
					log.Println("Login NOT successful!")
					emsg := "The password is not correct."
					c.HTML(http.StatusOK,
						"accounts.login_error.tmpl.html", emsg)
				}
			} else {
				log.Println("NO data from DB table.")
				emsg := "The username is not found."
				c.HTML(http.StatusOK,
					"accounts.login_error.tmpl.html", emsg)
			}

		}
	})

	//router.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.tmpl.html", nil)
	//
	//})

	router.GET("/", func(c *gin.Context) {
		authenticated, username := isAuth(c)
		if !authenticated {
			c.HTML(http.StatusOK, "index.tmpl.html",
				gin.H{"Anonymous": true, "Authenticated": false,
				"Username": ""})
		} else {
			rows, err := db.Query("SELECT ID, ChineseName, " +
				"EnglishName, Email, CellPhone, ShortPixName, " +
					" PictureURL, UsernameAsOwner " +
						" FROM Members order by ID DESC")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			members := make([]*Member, 0)
			for rows.Next() {
				member := new(Member)
				err := rows.Scan(&member.ID, &member.ChineseName,
					&member.EnglishName,
					&member.Email, &member.CellPhone,
						&member.ShortPixName, &member.PictureURL,
							&member.UsernameAsOwner)
				if err != nil {
					log.Fatal(err)
				}
				members = append(members, member)
			}
			for _, m := range members {
				if m.UsernameAsOwner == username {
					m.Disabled = false
				} else {
					m.Disabled = true
				}
			}
			c.HTML(http.StatusOK, "index_protected.tmpl.html",
				gin.H{"Anonymous": false, "Authenticated": true,
				"Username": username, "Members": members})

		}

	})

	router.Run(":" + port)
}
