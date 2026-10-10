package auth

import (
	"context"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/mailer"
)

var (
	ErrAlreadyVerified = errors.New("your email is already verified")
	ErrEmailTaken      = errors.New("that email is already used by another account")
)

// SendVerificationCode emails the player a fresh 6-digit code. newEmail,
// when set, first corrects the address (for a typo at sign-up).
func (s *Service) SendVerificationCode(ctx context.Context, userID, newEmail string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil || user.IsGuest {
		return errors.New("demo accounts don't have an email to verify")
	}
	if user.EmailVerifiedAt != nil {
		return ErrAlreadyVerified
	}
	email := user.Email
	if newEmail = strings.ToLower(strings.TrimSpace(newEmail)); newEmail != "" && newEmail != email {
		if err := validateEmail(newEmail); err != nil {
			return err
		}
		if other, _ := s.userRepo.GetByEmail(ctx, newEmail); other != nil {
			return ErrEmailTaken
		}
		if err := s.userRepo.SetEmail(ctx, userID, newEmail); err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return ErrEmailTaken
			}
			return err
		}
		email = newEmail
	}
	return s.sendCode(ctx, userID, email)
}

func (s *Service) sendCode(ctx context.Context, userID, email string) error {
	if s.mail == nil {
		return errors.New("email is not set up on this server")
	}
	code, err := newCode()
	if err != nil {
		return err
	}
	if err := s.codes.Replace(ctx, userID, email, code); err != nil {
		return err
	}
	return s.mail.Send(ctx, verificationEmail(email, code))
}

// VerifyEmail checks the code and marks the email verified.
func (s *Service) VerifyEmail(ctx context.Context, userID, code string) (*models.User, error) {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return nil, errors.New("enter the 6-digit code from the email")
	}
	email, err := s.codes.Check(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	ok, err := s.userRepo.MarkEmailVerified(ctx, userID, email)
	if err != nil {
		return nil, err
	}
	_ = s.codes.Delete(ctx, userID)
	if !ok {
		// Verified meanwhile, or the address changed after this code was sent.
		user, _ := s.userRepo.GetByID(ctx, userID)
		if user != nil && user.EmailVerifiedAt != nil {
			return s.withFlags(user), nil
		}
		return nil, errors.New("that code was for a different email — send a new one")
	}
	return s.GetUser(ctx, userID)
}

// validateEmail rejects malformed and throwaway addresses.
func validateEmail(email string) error {
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || !strings.Contains(domain, ".") || strings.ContainsAny(email, " \t<>,;") || len(email) > 254 {
		return errors.New("enter a valid email address")
	}
	if disposableDomains[domain] {
		return errors.New("please use a permanent email address, not a throwaway one")
	}
	return nil
}

// disposableDomains are common throwaway-inbox services.
var disposableDomains = map[string]bool{
	"mailinator.com": true, "guerrillamail.com": true, "guerrillamail.net": true, "sharklasers.com": true,
	"10minutemail.com": true, "10minutemail.net": true, "tempmail.com": true, "temp-mail.org": true,
	"temp-mail.io": true, "tempmail.net": true, "throwawaymail.com": true, "yopmail.com": true,
	"yopmail.net": true, "getnada.com": true, "nada.email": true, "dispostable.com": true,
	"trashmail.com": true, "maildrop.cc": true, "mintemail.com": true, "fakeinbox.com": true,
	"mailnesia.com": true, "emailondeck.com": true, "mohmal.com": true, "tempail.com": true,
	"moakt.com": true, "burnermail.io": true, "spamgourmet.com": true, "mailcatch.com": true,
	"tempr.email": true, "discard.email": true, "example.com": true, "test.com": true,
}

func verificationEmail(to, code string) mailer.Email {
	text := fmt.Sprintf("Your Mind Race verification code is %s\n\nIt expires in %d minutes. If you didn't sign up for Mind Race, you can ignore this email.\n\n— Mind Race · mindrace.in", code, int(codeTTL.Minutes()))
	htmlBody := fmt.Sprintf(`<!doctype html><html><body style="margin:0;background:#0b0e1f;font-family:Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#e8ebff">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#0b0e1f;padding:32px 12px"><tr><td align="center">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:440px;background:#141834;border:1px solid #2b3163;border-radius:14px;padding:32px 28px">
<tr><td style="font-size:22px;font-weight:800;letter-spacing:.04em">⚡ MIND<span style="color:#7c5cff">RACE</span></td></tr>
<tr><td style="padding-top:20px;font-size:15px;line-height:1.5;color:#c9cdf0">Here's your verification code:</td></tr>
<tr><td style="padding:16px 0"><div style="font-size:36px;font-weight:800;letter-spacing:10px;background:#0b0e1f;border:1px solid #2b3163;border-radius:10px;padding:14px 0;text-align:center;color:#ffffff">%s</div></td></tr>
<tr><td style="font-size:13.5px;line-height:1.5;color:#8b97ae">It expires in %d minutes. If you didn't sign up for Mind Race, you can ignore this email.</td></tr>
</table>
<p style="font-size:12px;color:#5d6788;margin-top:16px">Mind Race · <a href="https://mindrace.in" style="color:#7c5cff">mindrace.in</a></p>
</td></tr></table></body></html>`, html.EscapeString(code), int(codeTTL.Minutes()))
	return mailer.Email{To: to, Subject: code + " is your Mind Race code", Text: text, HTML: htmlBody}
}
