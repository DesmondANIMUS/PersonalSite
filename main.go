package gogosite

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"

	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/http2"
)

type userInfo struct {
	Name    string `bson:"name"`
	Message string `bson:"message"`
	Email   string `bson:"email"`
}

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("./*.htm"))

	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("./assets"))))
	
	server := new(http.Server)
	h2Conf := new(http2.Server)
	http2.ConfigureServer(server, h2Conf)

	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	http.HandleFunc("/contact", contact)
	http.HandleFunc("/CV", cv)
}

func cv(w http.ResponseWriter, r *http.Request) {
	streamPDFbytes, err := ioutil.ReadFile("./assets/CV.pdf")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	b := bytes.NewBuffer(streamPDFbytes)

	// stream straight to client(browser)
	w.Header().Set("Content-type", "application/pdf")

	if _, err := b.WriteTo(w); err != nil { // <----- here!
		fmt.Fprintf(w, "%s", err)
	}

	log.Println(r.URL.Path)
}

func index(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "index.htm", nil)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Println(r.URL.Path)
}

func about(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "about.htm", nil)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Println(r.URL.Path)
}

func contact(w http.ResponseWriter, r *http.Request) {
	var use userInfo
	response := "Failed :("

	if r.Method == http.MethodPost {

		use.Name = r.FormValue("name")
		use.Message = r.FormValue("message")
		use.Email = r.FormValue("email")

		if use.Name == "" || use.Message == "" || use.Email == "" {
			response = "not getting data :("
		} else {
			err := sendMail(use.Message, use.Email)
			if err != nil {
				response = "could not send email :("
			} else {
				if err != nil {
					response = "Failed :("
				} else {
					response = "Success"
				}
			}
		}
	}

	if response == "Success" {

		id, _ := uuid.NewV4()

		http.SetCookie(w, &http.Cookie{
			Name:  "This-Session-Cookie",
			Value: id.String(),
		})
		http.Redirect(w, r, "/", 301)
	}

	_, err := r.Cookie("This-Session-Cookie")
	if err != nil {
		err = tpl.ExecuteTemplate(w, "contact.htm", nil)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		http.Redirect(w, r, "/", 301)
	}

	fmt.Println(err)
	log.Println(r.URL.Path)
}

func sendMail(Body, to string) error {
	from := "0incognitogaurav0@gmail.com"
	password := "desmond_ANIMUS12"

	msg := "From: " + to + "\r\n" +
		"To: " + from + "\r\n" +
		"MIME-Version: 1.0" + "\r\n" +
		"Content-type: text/html" + "\r\n" +
		"Subject: Reigstration Success" + "\r\n\r\n" +
		Body + "\r\n"

	err := smtp.SendMail("smtp.gmail.com:587", smtp.PlainAuth("", from, password, "smtp.gmail.com"), from, []string{to}, []byte(msg))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Verification Message Sent")
	return nil
}
