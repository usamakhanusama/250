@echo off
echo ==========================================================
echo    SUST CSE 250 - Usama's Git Push Automation Script
echo ==========================================================

if not exist ".git" (
    echo Error: Please run this script from the root of your cloned proj250 git repository!
    pause
    exit /b 1
)

echo 1. Configuring Git identity for usamakhanusama...
git config user.name "usamakhanusama"
set "usama_email=editingshop005@gmail.com"
git config user.email "%usama_email%"
echo Git email successfully set to: %usama_email%

echo 2. Creating feature branch...
git checkout -b feature/usama_auth_student_result 2>nul || git checkout feature/usama_auth_student_result

echo 3. Staging Usama's Go packages...
git add backend/internal/auth backend/internal/student backend/internal/result backend/cmd/server/main.go

echo 4. Committing as Usama...
git commit -m "feat(backend): implement JWT auth, student section assigner, and CGPA engine by Usama (2023331088)"

echo 5. Pushing to GitHub...
git push -u origin feature/usama_auth_student_result

echo ==========================================================
echo SUCCESS! Your Go packages have been pushed to GitHub!
echo ==========================================================
pause