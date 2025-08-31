#!/bin/bash
cd frontend/
npm ci
npm run build
cd ../backend/
CGO_ENABLED='1' go build -tags sqlite -o ../pixel-protocol .
