#!/bin/bash
set -e

echo "=========================================================="
echo "   SUST CSE 250 - Usama's Git Push Automation Script"
echo "=========================================================="

# Check if inside git repo
if [ ! -d ".git" ]; then
    echo "❌ Error: Please run this script from the root directory of your cloned proj250 git repository!"
    exit 1
fi

echo "1. Configuring Git identity for Usamakhanusama..."
git config user.name "usamakhanusama"
usama_email="editingshop005@gmail.com"
git config user.email "$usama_email"
echo "Git email successfully set to: $usama_email"

echo "2. Creating feature branch: feature/usama-auth-student-result..."
git checkout -b feature/usama-auth-student-result 2>/dev/null || git checkout feature/usama-auth-student-result

echo "3. Staging Usama's Go modules (auth, student, result, main.go)..."
git add backend/internal/auth backend/internal/student backend/internal/result backend/cmd/server/main.go

echo "4. Running Go unit tests to verify correctness..."
cd backend
go test -v ./internal/auth ./internal/student ./internal/result
cd ..

echo "5. Committing under Usama's name..."
git commit -m "feat(backend): implement JWT auth, student section assigner, and CGPA engine by Usama (2023331088)"

echo "6. Pushing to GitHub..."
git push -u origin feature/usama-auth-student-result

echo "=========================================================="
echo "🎉 SUCCESS! Your Go modules have been pushed from your account!"
echo "Now go to GitHub and create a Pull Request to merge into main."
echo "=========================================================="