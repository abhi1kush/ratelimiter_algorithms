Rate Limiter in Golang

📌 Overview

This project implements a rate limiter in Golang, inspired by Alex Xu's "System Design Interview" book. Rate limiting is essential for controlling API access, preventing abuse, and ensuring fair resource allocation.

🚀 Features

✅ Token Bucket Algorithm: Efficient rate-limiting mechanism.

✅ Fixed Window Algorithm: Simple implementation with a reset period.

✅ Sliding Window Log Algorithm: Provides more accurate rate-limiting decisions.

✅ Sliding Window Counter Algorithm: Acheives more accurate rate-limiting decisions by considering previous window.

🛠 Installation

🧪 Testing

Run unit tests:

$ go test ./...

