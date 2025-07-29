#!/bin/bash

mkdir -p static/css
mkdir -p static/js
mkdir -p static/webfonts
mkdir -p static/fonts

echo "Downloading assets..."

# Download Tailwind CSS
curl -sL "https://cdn.tailwindcss.com" -o "static/js/tailwindcss.js"

# Download Font Awesome CSS and its webfonts
curl -sL "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css" -o "static/css/all.min.css"
curl -sL "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/webfonts/fa-brands-400.woff2" -o "static/webfonts/fa-brands-400.woff2"
curl -sL "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/webfonts/fa-regular-400.woff2" -o "static/webfonts/fa-regular-400.woff2"
curl -sL "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/webfonts/fa-solid-900.woff2" -o "static/webfonts/fa-solid-900.woff2"
curl -sL "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/webfonts/fa-v4compatibility.woff2" -o "static/webfonts/fa-v4compatibility.woff2"

# Update Font Awesome CSS to use local webfonts path
sed -i.bak 's|https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/webfonts/|/static/webfonts/|g' static/css/all.min.css
rm static/css/all.min.css.bak

# Download Inter font CSS from Google Fonts
curl -sL "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" -A "Mozilla/5.0" -o "static/css/inter.css"

# Download font files referenced in the CSS
grep -o 'https://fonts.gstatic.com/s/inter/[^)]*' static/css/inter.css | while read -r url; do
  # Remove the single quote from the end
  clean_url=$(echo "$url" | sed "s/'$//")
  filename=$(basename "$clean_url")
  curl -sL "$clean_url" -o "static/fonts/$filename"
done

# Update font CSS to use local font files
sed -i.bak 's|https://fonts.gstatic.com/s/inter/v[0-9]*/|/static/fonts/|g' static/css/inter.css
rm static/css/inter.css.bak

echo "All assets downloaded successfully!"
