FROM node:20-slim

WORKDIR /app
COPY package.json ./
RUN npm install
COPY slides.md ./
COPY public/ ./public/
RUN npx slidev build --base ./ --out slides
