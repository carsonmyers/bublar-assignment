package v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/gbrlsnchs/jwt/v2"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

type contextKey string

func (c contextKey) String() string {
	return "ORLv1:" + string(c)
}

var (
	ctxToken = contextKey("Token")
	ctxRID   = contextKey("rID")
	ctxLog   = contextKey("logger")
)

// GetAuth - Get the authentication information from the request
func GetAuth(r *http.Request) *jwt.JWT {
	if token, ok := r.Context().Value(ctxToken).(*jwt.JWT); ok {
		return token
	}

	return nil
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqLog := GetLogger(r)

		var token jwt.JWT
		cookie, err := r.Cookie("AUTH")
		if err != nil {
			if err != http.ErrNoCookie {
				reqLog.Error("Error checking for auth cookie", zap.Error(err))
			} else {
				reqLog.Debug("No auth cookie present")
			}
		} else {
			payload, err := base64.StdEncoding.DecodeString(cookie.Value)
			if err != nil {
				reqLog.Error("Error decoding auth cookie", zap.String("cookie", cookie.String()), zap.Error(err))
			}

			if err := json.Unmarshal(payload, &token); err != nil {
				reqLog.Error("Error unmarshaling auth cookie", zap.String("token", string(payload)), zap.Error(err))
			}

			expires := time.Unix(token.ExpirationTime, 0).Sub(time.Now())
			reqLog.Debug("Authorized request", zap.String("username", token.Audience), zap.Duration("expiresIn", expires))
		}

		ctx := context.WithValue(r.Context(), ctxToken, &token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetAuth - set the auth cookie to a JWT token
func SetAuth(w http.ResponseWriter, r *http.Request, token *jwt.JWT) *errors.Error {
	reqLog := GetLogger(r)

	payload, err := json.Marshal(token)
	if err != nil {
		reqLog.Error("Error marshalling auth token", zap.String("iss", token.Issuer), zap.String("aud", token.Audience), zap.Error(err))
		return errors.EInternal.NewErrorf("failed to marshal auth token").Wrap(err)
	}

	encoded := base64.StdEncoding.EncodeToString(payload)

	cookie := &http.Cookie{
		Name:  "AUTH",
		Value: encoded,
	}

	reqLog.Info("Set auth cookie", zap.String("username", token.Audience), zap.String("value", encoded))
	http.SetCookie(w, cookie)
	return nil
}

// GetLogger - Get the request logger
func GetLogger(r *http.Request) *zap.Logger {
	if log, ok := r.Context().Value(ctxLog).(*zap.Logger); ok {
		return log
	}

	return logger.GetLogger()
}

// RIDMiddleware - middleware that attaches a ULID to each request
func RIDMiddleware() func(http.Handler) http.Handler {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var rID string
			if existing := r.Header.Get("Request-ID"); existing != "" {
				rID = existing
			} else {
				rID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
				r.Header.Set("Request-ID", rID)
			}

			ctx := context.WithValue(r.Context(), ctxRID, rID)
			next.ServeHTTP(w, r.WithContext(ctx))

			if afterRID, ok := r.Context().Value(ctxRID).(string); ok {
				if afterRID != rID {
					GetLogger(r).Warn("Request ID changed during processing", zap.String("before", rID), zap.String("after", afterRID))
				}
			}
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	status     string
	bytes      int
}

func lrw(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, 200, http.StatusText(200), 0}
}

func (l *loggingResponseWriter) WriteHeader(code int) {
	l.statusCode = code
	l.status = fmt.Sprintf("%d %s", code, http.StatusText(code))
	l.ResponseWriter.WriteHeader(code)
}

func (l *loggingResponseWriter) Write(data []byte) (bytes int, err error) {
	bytes, err = l.ResponseWriter.Write(data)
	if err != nil {
		return
	}

	l.bytes += bytes
	return
}

// LoggingMiddleware - middleware which logs every request and response
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()

		if id, ok := r.Context().Value(ctxRID).(string); ok {
			log = log.With(zap.String("requestID", id))
		}

		ctx := context.WithValue(r.Context(), ctxLog, log)

		log.Info("==> HTTP", zap.String("method", r.Method), zap.String("uri", r.RequestURI))

		writer := lrw(w)
		start := time.Now()

		next.ServeHTTP(writer, r.WithContext(ctx))

		d := time.Now().Sub(start)
		log.Info(fmt.Sprintf("<== %s", writer.status), zap.Int("status", writer.statusCode), zap.Int("bytes", writer.bytes), zap.Duration("duration", d))
	})
}
