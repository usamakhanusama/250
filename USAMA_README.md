# SUST CSE 250 — Usama's Go Backend Contribution Package
## Team Member: Usama (Registration: 2023331088 | Section B)

উসামা, এই জিপ ফাইলে তোমার অংশের সম্পূর্ণ **Go (Golang)** ব্যাকএন্ড কোড রয়েছে। তুমি তোমার নিজস্ব গিটহাব অ্যাকাউন্ট থেকে এই কোড পুশ করবে।

---

### ১. তোমার প্যাকেজসমূহ (মোট ১,৩১০ লাইন Go কোড ও ৩টি প্যাকেজ)

1. **`internal/auth/` (৭টি ফাইল, ৭৮১ লাইন)**
   - `jwt.go`, `password.go`, `service.go`, `handler.go`, `models.go`
   - `jwt_test.go`, `password_test.go`
   - **কাজ:** Argon2id ক্রিপ্টোগ্রাফিক পাসওয়ার্ড হ্যাশিং, HMAC-SHA256 সাইনড JWT টোকেন ইস্যু, রিফ্রেশ টোকেন ও সেশন সিকিউরিটি।

2. **`internal/student/` (৪টি ফাইল, ২৪৭ লাইন)**
   - `models.go`, `service.go`, `handler.go`, `student_test.go`
   - **কাজ:** ব্যাচ ২৩ স্টুডেন্ট ডিরেক্টরি এবং রোল নম্বরের শেষ দুই ডিজিট জোড় হলে সেকশন B এবং বিজোড় হলে সেকশন A অ্যাসাইন করার অটোমেশন ও ব্যাচ অ্যানালিটিক্স।

3. **`internal/result/` (৪টি ফাইল, ২৮২ লাইন)**
   - `models.go`, `service.go`, `handler.go`, `result_test.go`
   - **কাজ:** সেমিস্টার ১, ২ ও ৩-এর ট্রান্সক্রিপ্ট, ৪.০ স্কেলে কিউমুলেটিভ CGPA ক্যালকুলেশন ইঞ্জিন ও **Target GPA Simulator** ("What-If" অ্যানালাইসিস)।

4. **`cmd/server/main.go`**
   - সার্ভার এন্ট্রি পয়েন্টে তোমার তিনটি মডিউলের রাউট রেজিস্টার করা আছে:
     - `POST /api/v1/auth/login`
     - `POST /api/v1/auth/refresh`
     - `GET  /api/v1/students`
     - `GET  /api/v1/students/{reg}`
     - `GET  /api/v1/students/analytics`
     - `GET  /api/v1/results/transcript/{reg}`
     - `GET  /api/v1/results/simulate`

---

### ২. যেভাবে সহজে তোমার গিটহাব থেকে পুশ করবে (Step-by-Step)

#### নিয়ম (Option A — ১-ক্লিক অটো স্ক্রিপ্ট):
- তুমি যদি **Mac বা Linux** ব্যবহার করো:
  টার্মিনালে এই কমান্ডটি চালাও:
  ```bash
  chmod +x PUSH_TO_GITHUB.sh
  ./PUSH_TO_GITHUB.sh
  ```
- তুমি যদি **Windows** ব্যবহার করো:
  `PUSH_TO_GITHUB.bat` ফাইলে ডাবল-ক্লিক করো অথবা CMD-তে চালাও।

---

#### ম্যানুয়াল নিয়ম (Option B — নিজে গিট কমান্ড দিয়ে):
১. প্রজেক্টের রুট ডিরেক্টরিতে টার্মিনাল ওপেন করো।
২. তোমার গিট আইডেন্টিটি সেট করো (যাতে গিটহাবে তোমার নাম ও ছবি আসে):
   ```bash
   git config user.name "Usama"
   git config user.email "usama@student.sust.edu"
   ```
৩. একটি নতুন ব্রাঞ্চ খোলো:
   ```bash
   git checkout -b feature/usama-auth-student-result
   ```
৪. এই জিপের `backend/` ফোল্ডারটি তোমার প্রজেক্টের `backend/` ফোল্ডারে কপি/পেস্ট করে দাও (Replace any matching files)।
৫. টেস্ট চালিয়ে দেখে নাও সব পাস করে কিনা:
   ```bash
   cd backend
   go test -v ./...
   cd ..
   ```
৬. স্টেজ ও কমিট করো:
   ```bash
   git add backend/internal/auth backend/internal/student backend/internal/result backend/cmd/server/main.go
   git commit -m "feat(backend): implement JWT auth, student section assigner, and CGPA engine by Usama (2023331088)"
   ```
৭. গিটহাবে পুশ করে দাও:
   ```bash
   git push -u origin feature/usama-auth-student-result
   ```
   (অথবা তোমার মেইন ব্রাঞ্চে মার্জ করে পুশ করো)।

---

### ৩. ডিফেন্সে স্যারকে যা বলবে (Viva Talking Points)
> *"স্যার, আমি Go ব্যাকএন্ডে ৩টি কোর প্যাকেজ তৈরি করেছি: `internal/auth`, `internal/student`, এবং `internal/result`। আমি Argon2id সিকিউরিটি ও JWT অথেনটিকেশন ইঞ্জিন তৈরি করেছি, স্টুডেন্টদের রোল নাম্বারের প্যারিটি দিয়ে স্বয়ংক্রিয় সেকশন B (ইভেন) এবং A (অড) ভাগ করার অ্যালগরিদম লিখেছি এবং ৪.০ স্কেলে CGPA ক্যালকুলেশন ও টার্গেট GPA প্রেডিকশন সিমুলেটর ইমপ্লিমেন্ট করেছি। সাথে আমার প্রতিটি প্যাকেজের জন্য ইউনিট টেস্ট লিখেছি যা ১০০% পাস করেছে।*"
