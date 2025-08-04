#!/bin/bash

echo "🚀 Testing React Router v7 Migration"
echo "======================================"

# Check if required files exist
echo "✅ Checking required files..."
required_files=(
    "index.html"
    "vite.config.ts"
    "src/main.tsx"
    "src/routes/layout.tsx"
    "src/routes/app.tsx"
    "src/routes/survey.\$urlSlug.tsx"
    "src/routes/app.surveys.\$surveyUuid.responses.tsx"
    "src/routes/not-found.tsx"
    "package.json.new"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "   ✓ $file"
    else
        echo "   ✗ $file (missing)"
        exit 1
    fi
done

echo ""
echo "✅ Checking environment setup..."
if [ -f ".env.example" ]; then
    echo "   ✓ .env.example exists"
    if grep -q "VITE_API_URL" .env.example; then
        echo "   ✓ VITE_API_URL configured"
    else
        echo "   ✗ VITE_API_URL not found in .env.example"
        exit 1
    fi
else
    echo "   ✗ .env.example missing"
    exit 1
fi

echo ""
echo "✅ Checking TypeScript configuration..."
if grep -q "vite/client" tsconfig.json; then
    echo "   ✓ Vite types configured"
else
    echo "   ✗ Vite types not configured"
    exit 1
fi

echo ""
echo "🎉 Migration verification complete!"
echo ""
echo "Next steps:"
echo "1. mv package.json package.json.old"
echo "2. mv package.json.new package.json"
echo "3. npm install"
echo "4. cp .env.example .env"
echo "5. npm run dev"