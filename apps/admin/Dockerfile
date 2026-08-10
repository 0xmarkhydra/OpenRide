FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS build
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
RUN addgroup -S flashx && adduser -S -G flashx -u 10001 flashx
COPY --from=build --chown=flashx:flashx /app/.next/standalone ./
COPY --from=build --chown=flashx:flashx /app/.next/static ./.next/static
USER flashx
EXPOSE 3000
CMD ["node", "server.js"]
