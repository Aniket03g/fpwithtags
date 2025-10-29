package log

// Example usage of the structured logger
//
// Basic logging:
//   import "backend/internal/log"
//
//   log.Info("Application started")
//   log.Debugf("Processing request for user: %s", userID)
//   log.Error("Failed to connect to database")
//
// Structured logging with fields:
//   log.WithFields(logrus.Fields{
//       "user_id": 123,
//       "action": "login",
//       "ip": "192.168.1.1",
//   }).Info("User logged in")
//
// Logging with error context:
//   if err != nil {
//       log.WithError(err).Error("Failed to process request")
//   }
//
// Using the logger instance directly:
//   logger := log.Logger()
//   logger.WithField("module", "auth").Info("Authentication successful")
//
// Setting log level (typically in main.go):
//   log.SetLevel(logrus.DebugLevel)
//
// Environment variable configuration:
//   Set LOG_LEVEL environment variable to: debug, info, warn, or error
//   Example: LOG_LEVEL=debug go run main.go
//
// Log output format (JSON):
//   {
//     "level": "info",
//     "msg": "User logged in",
//     "time": "2025-01-15 14:30:45",
//     "user_id": 123,
//     "action": "login",
//     "ip": "192.168.1.1"
//   }
//
// Logs are written to:
//   - stdout (console)
//   - /var/log/featureplus/app.log (file)
//
// In Docker, logs appear in the host's ./logs/ directory
