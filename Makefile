.PHONY: icons clean

# Regenerate all application icons from source image
# Source: ~/Desktop/appiconraw.png (single source of truth)
icons:
	python3 tools/icongen/main.py

# Regenerate with custom source
icons-from:
	@if [ -z "$(SRC)" ]; then echo "Usage: make icons-from SRC=path/to/image.png"; exit 1; fi
	python3 tools/icongen/main.py --source "$(SRC)"

clean:
	rm -f build/appicon.png
	rm -f build/bin/SwiftRay.app/Contents/Resources/iconfile.icns
	rm -f build/windows/icon.ico
	rm -rf build/icons/
	rm -f frontend/src/assets/images/logo.png
