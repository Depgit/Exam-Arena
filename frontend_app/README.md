# Exam Arena Android App

A modern, competitive Android application built with **Kotlin**, **Jetpack Compose**, and **Material 3 Expressive**, designed to compete in live real-time exam battles, track comprehensive performance analytics, and climb global leaderboards.

---

## Architecture Overview

```
frontend_app/
├── app/
│   └── src/main/java/com/examarena/
│       ├── ExamArenaApplication.kt     # App Entry & DI Container
│       ├── MainActivity.kt             # Edge-to-edge Root Activity
│       ├── ExamArenaApp.kt             # Root Scaffold & Dynamic Bottom Navigation
│       │
│       ├── core/
│       │   ├── common/                 # Resource<T> & Result sealed classes
│       │   ├── datastore/              # SessionManager (JWT Token & Preferences)
│       │   ├── designsystem/           # Theme, Colors, Typography, Shapes & Components
│       │   │   ├── components/         # AppButton, AppCard, AppTextField, AppAvatar, etc.
│       │   │   ├── motion/             # MotionTokens & Navigation transitions
│       │   │   └── theme/              # Color, Type, Shape & Theme palettes
│       │   ├── di/                     # Lightweight AppContainer DI
│       │   ├── navigation/             # Type-safe Screen routes & NavGraph
│       │   ├── network/                # ApiClient with OkHttp & Kotlinx Serialization
│       │   └── websocket/              # Realtime WebSocket client & event hub
│       │
│       ├── data/
│       │   ├── model/                  # Data Transfer Objects (Auth, Match, Category, Leaderboard, Practice)
│       │   └── repository/             # Auth, Match, Leaderboard, Practice, and User Repositories
│       │
│       └── feature/
│           ├── splash/                 # Splash screen & session check
│           ├── onboarding/             # 3-step feature onboarding carousel
│           ├── auth/                   # Login, Register, and Instant Guest Play
│           ├── home/                   # Dashboard, Quick Match CTA, Category selectors
│           ├── play/                   # Play setup, 1v1 ranked & 6-char Friendly Rooms
│           ├── matchmaking/            # Radar search animation & ELO matching
│           ├── matchfound/             # 1v1 VS layout & 3-2-1 countdown sequence
│           ├── battle/                 # Timed live competition HUD & real-time scoring
│           ├── result/                 # Victory celebration, score tally & rating deltas
│           ├── leaderboard/            # Category leaderboards & Top 3 Podium
│           ├── practice/               # Solo study mode with instant solution explanations
│           └── profile/                # Profile statistics, ratings & server configuration
```

---

## Key Features & User Flows

1. **Authentication Flow**:
   - Register new account with validation.
   - Login with email or username.
   - Instant 1-tap **Guest Play** for instant evaluation without typing credentials.
2. **Realtime 1v1 Matchmaking**:
   - Enqueue in ranked matches for any exam category (`SSC CGL`, `Banking / IBPS`, `Railways`, `UPSC`, `CAT / MBA`, `State PSC`).
   - Animated radar search waves with expanding rating buckets ($\pm 100 \to \pm 250$).
3. **Live Multiplayer Battle Engine**:
   - Live WebSocket communication (`ws://10.0.2.2:8080/ws?token=<JWT>`).
   - Timed matches (120s) with live score synchronization and opponent answer status indicators.
   - Authoritative server evaluation (preventing client-side tampering).
4. **Friendly Room Codes**:
   - Generate unique 6-character room codes to challenge friends anywhere.
5. **Solo Study Practice Mode**:
   - Interactive question practice with instant explanations and category mastery.
6. **Leaderboards**:
   - Visual Top 3 podium (🥇 Gold, 🥈 Silver, 🥉 Bronze) and rank listings.
7. **Profile & Dynamic Server Configuration**:
   - Switch between Android Emulator (`10.0.2.2:8080`), Localhost (`localhost:8080`), or Custom LAN IP from the app settings dialog.

---

## Building and Running

### Prerequisites
- **Android Studio** (Koala / Ladybug / Meerkat or newer) with Android SDK Platform 35.
- **JDK 21** (bundled with Android Studio at `Android Studio/jbr`).

### Opening in Android Studio
1. Open Android Studio.
2. Select **Open** and choose the `frontend_app` directory.
3. Allow Gradle to sync dependencies.
4. Run the Go backend (`cd backend && go run ./cmd/server`).
5. Launch the Android app on an Emulator or connected Android device.
