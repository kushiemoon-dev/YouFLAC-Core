ANDROID_NDK ?= /opt/android-sdk/ndk/29.0.14206865
NDK_TOOLCHAIN = $(ANDROID_NDK)/toolchains/llvm/prebuilt/linux-x86_64

OUT_DIR ?= build
MOBILE_DIR ?= ../YouFLAC-Mobile

android: android-arm64 android-arm android-x86_64

android-arm64:
	@mkdir -p $(OUT_DIR)/android/arm64-v8a
	CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=$(NDK_TOOLCHAIN)/bin/aarch64-linux-android21-clang go build -buildmode=c-shared -o $(OUT_DIR)/android/arm64-v8a/libyouflac.so ./cmd/bridge/

android-arm:
	@mkdir -p $(OUT_DIR)/android/armeabi-v7a
	CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 CC=$(NDK_TOOLCHAIN)/bin/armv7a-linux-androideabi21-clang go build -buildmode=c-shared -o $(OUT_DIR)/android/armeabi-v7a/libyouflac.so ./cmd/bridge/

android-x86_64:
	@mkdir -p $(OUT_DIR)/android/x86_64
	CGO_ENABLED=1 GOOS=android GOARCH=amd64 CC=$(NDK_TOOLCHAIN)/bin/x86_64-linux-android21-clang go build -buildmode=c-shared -o $(OUT_DIR)/android/x86_64/libyouflac.so ./cmd/bridge/

install-android: android
	@for abi in arm64-v8a armeabi-v7a x86_64; do mkdir -p $(MOBILE_DIR)/android/app/src/main/jniLibs/$$abi; cp $(OUT_DIR)/android/$$abi/libyouflac.so $(MOBILE_DIR)/android/app/src/main/jniLibs/$$abi/; done
	@echo "Installed .so files into $(MOBILE_DIR)"

host:
	@mkdir -p $(OUT_DIR)/host
	CGO_ENABLED=1 go build -buildmode=c-shared -o $(OUT_DIR)/host/libyouflac.so ./cmd/bridge/

IOS_SDK_PATH ?= $(shell xcrun --sdk iphoneos --show-sdk-path 2>/dev/null)
IOS_CLANG ?= $(shell xcrun --sdk iphoneos --find clang 2>/dev/null || echo clang)
IOS_MIN_VERSION ?= 16.0

ios:
	@mkdir -p $(OUT_DIR)/ios
	@if [ -z "$(IOS_SDK_PATH)" ]; then echo "ERROR: iOS SDK not found."; exit 1; fi
	CGO_ENABLED=1 GOOS=ios GOARCH=arm64 CC="$(IOS_CLANG) -arch arm64 -isysroot $(IOS_SDK_PATH) -miphoneos-version-min=$(IOS_MIN_VERSION)" go build -buildmode=c-archive -o $(OUT_DIR)/ios/libyouflac.a ./cmd/bridge/

install-ios: ios
	@mkdir -p $(MOBILE_DIR)/ios/Runner
	cp $(OUT_DIR)/ios/libyouflac.a $(MOBILE_DIR)/ios/Runner/
	cp $(OUT_DIR)/ios/libyouflac.h $(MOBILE_DIR)/ios/Runner/
	@echo "Installed .a into $(MOBILE_DIR)/ios/Runner/"

test:
	go test ./... -v -timeout 120s

clean:
	rm -rf $(OUT_DIR)

.PHONY: android android-arm64 android-arm android-x86_64 install-android host ios install-ios test clean
