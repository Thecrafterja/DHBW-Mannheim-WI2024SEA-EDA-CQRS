package main

import (
	"html/template"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

var amqpURI = "amqp://guest:guest@localhost:5672/"
var publisher message.Publisher

const htmlTemplate = `
<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Message Publisher</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        h1 {
            color: #333;
            text-align: center;
        }
        form {
            display: flex;
            flex-direction: column;
        }
        label {
            margin-bottom: 8px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"], textarea {
            padding: 10px;
            margin-bottom: 15px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-family: Arial, sans-serif;
            font-size: 14px;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
        }
        input[type="submit"] {
            padding: 12px;
            background-color: #4CAF50;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            font-weight: bold;
        }
        input[type="submit"]:hover {
            background-color: #45a049;
        }
        .message {
            margin-top: 20px;
            padding: 15px;
            border-radius: 4px;
            text-align: center;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Message Publisher</h1>
        <form method="POST" action="/publish">
            <label for="message">Nachricht:</label>
            <textarea id="message" name="message" required placeholder="Geben Sie Ihre Nachricht hier ein..."></textarea>
            <input type="submit" value="Senden">
        </form>
        {{if .Message}}
            <div class="message {{.MessageType}}">
                {{.Message}}
            </div>
        {{end}}
    </div>
</body>
</html>
`

type TemplateData struct {
	Message     string
	MessageType string
}

func main() {
	var err error
	amqpConfig := amqp.NewDurableQueueConfig(amqpURI)

	publisher, err = amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(false, false))
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/publish", handlePublish)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Template-Fehler", http.StatusInternalServerError)
		return
	}

	data := TemplateData{}
	tmpl.Execute(w, data)
}

func handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	topic := "example-topic"
	messageText := r.FormValue("message")

	if messageText == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	msg := message.NewMessage(watermill.NewUUID(), []byte(messageText))
	err := publisher.Publish(topic, msg)
	if err != nil {
		println("Nachricht konnte nicht versendet werden")
		return
	}

	tmpl, _ := template.New("success").Parse(htmlTemplate)
	data := TemplateData{
		Message:     "Nachricht versendet",
		MessageType: "success",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
