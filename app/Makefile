.PHONY: init

# Run Android app with hot reload
run-android: android/app/google-services.json
	flutter run --hot

# initial setup
init:
	@echo "==> Generating key pair"
	keytool -genkeypair
