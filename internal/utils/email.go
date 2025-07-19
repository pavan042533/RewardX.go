package utils
import(
	"fmt"
	"gopkg.in/gomail.v2"
	"os"
)

func SendOTPEmail(to, otp string) error {
	m := gomail.NewMessage()

	// Sender email
	from := os.Getenv("SMTP_USERNAME")

	// Set email headers
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your OTP Code from RewardX")

	// Email body
	body := fmt.Sprintf("%s is your RewardX verification OTP. Please do not share it with anyone.\n It will expire in 5 minutes.\n Team RewardX", otp)
	m.SetBody("text/plain", body)

	// Mail server config
	port := 587
	dailer := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		from,
		os.Getenv("SMTP_PASSWORD"),
	)

	// Send the mail
	if err := dailer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
