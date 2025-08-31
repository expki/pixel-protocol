#!/bin/bash
cd backend/
CGO_ENABLED='1' go run -tags sqlite .
