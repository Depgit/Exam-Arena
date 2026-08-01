import { 
  db, 
  auth, 
  signInWithPopup, 
  GoogleAuthProvider, 
  signInAnonymously as firebaseSignInAnonymously, 
  signOut as firebaseSignOut,
  onAuthStateChanged,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  sendPasswordResetEmail,
  updateProfile,
  User 
} from './firebase';
import { 
  doc, 
  getDoc, 
  setDoc, 
  updateDoc, 
  collection, 
  onSnapshot, 
  query, 
  orderBy, 
  limit, 
  addDoc, 
  where,
  deleteDoc
} from 'firebase/firestore';
import { PlayerProfile, ChatMessage, LeaderboardEntry, MatchState, AppNotification, MatchPlayer, SEARCHING_MATCH_PLAYER } from './types';
import { INITIAL_USER_PROFILE } from './storage';

// ═══════════════════════════════════════════════════════════════════════════
// 1. FIREBASE AUTHENTICATION SERVICES
// ═══════════════════════════════════════════════════════════════════════════

/** Google Authentication */
export async function signInWithGoogle(): Promise<User | null> {
  try {
    const provider = new GoogleAuthProvider();
    const result = await signInWithPopup(auth, provider);
    return result.user;
  } catch (error) {
    console.error('Error signing in with Google:', error);
    throw error;
  }
}

/** Email & Password Sign Up */
export async function signUpWithEmailAndPassword(
  email: string, 
  password: string, 
  username: string
): Promise<User> {
  try {
    const res = await createUserWithEmailAndPassword(auth, email, password);
    if (res.user) {
      await updateProfile(res.user, { displayName: username });
    }
    return res.user;
  } catch (error) {
    console.error('Error creating user with email/password:', error);
    throw error;
  }
}

/** Email & Password Sign In */
export async function signInWithEmailAndPasswordUser(
  email: string, 
  password: string
): Promise<User> {
  try {
    const res = await signInWithEmailAndPassword(auth, email, password);
    return res.user;
  } catch (error) {
    console.error('Error signing in with email/password:', error);
    throw error;
  }
}

/** Password Reset Email */
export async function sendResetPasswordEmail(email: string): Promise<void> {
  try {
    await sendPasswordResetEmail(auth, email);
  } catch (error) {
    console.error('Error sending password reset email:', error);
    throw error;
  }
}

/** Anonymous Duelist Authentication */
export async function signInAnonymouslyUser(): Promise<User | null> {
  try {
    const result = await firebaseSignInAnonymously(auth);
    return result.user;
  } catch (error) {
    console.error('Error signing in anonymously:', error);
    throw error;
  }
}

/** Sign Out */
export async function logOutUser(): Promise<void> {
  try {
    if (auth.currentUser) {
      await setUserStatusInFirestore(auth.currentUser.uid, 'offline');
    }
    await firebaseSignOut(auth);
  } catch (error) {
    console.error('Error signing out:', error);
    throw error;
  }
}

/** Auth State Listener with automatic profile creation & sync */
export function listenToAuthAndProfile(
  onProfileLoaded: (profile: PlayerProfile, user: User | null) => void
) {
  return onAuthStateChanged(auth, async (user) => {
    if (!user) {
      try {
        await firebaseSignInAnonymously(auth);
      } catch (err) {
        console.warn('Auto anonymous authentication failed:', err);
        onProfileLoaded(INITIAL_USER_PROFILE, null);
      }
      return;
    }

    const userDocRef = doc(db, 'users', user.uid);
    
    // Subscribe to live profile changes
    const unsubscribeProfile = onSnapshot(
      userDocRef, 
      async (snapshot) => {
        if (snapshot.exists()) {
          const data = snapshot.data() as PlayerProfile;
          onProfileLoaded({ ...data, id: user.uid }, user);
        } else {
          const newProfile: PlayerProfile = {
            ...INITIAL_USER_PROFILE,
            id: user.uid,
            username: user.displayName || `Duelist_${user.uid.slice(0, 5)}`,
            avatarUrl: user.photoURL || INITIAL_USER_PROFILE.avatarUrl,
            joinedDate: new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' }),
          };
          try {
            await setDoc(userDocRef, { ...newProfile, updatedAt: Date.now() });
            onProfileLoaded(newProfile, user);
          } catch (err) {
            console.error('Failed to create user doc:', err);
            onProfileLoaded(newProfile, user);
          }
        }
      },
      (error) => {
        console.warn('Profile listener snapshot note:', error?.message || error);
        onProfileLoaded(INITIAL_USER_PROFILE, user);
      }
    );

    return () => unsubscribeProfile();
  });
}

// ═══════════════════════════════════════════════════════════════════════════
// 2. FIRESTORE TABLE: 'users' (User Profiles, Stats, & Leaderboards)
// ═══════════════════════════════════════════════════════════════════════════

/** Get Single User Profile */
export async function getUserProfileFromFirestore(uid: string): Promise<PlayerProfile | null> {
  try {
    const userRef = doc(db, 'users', uid);
    const snap = await getDoc(userRef);
    if (snap.exists()) {
      return { id: snap.id, ...snap.data() } as PlayerProfile;
    }
    return null;
  } catch (e) {
    console.error('Error fetching user profile:', e);
    return null;
  }
}

/** Save or Update User Profile */
export async function saveProfileToFirestore(profile: PlayerProfile): Promise<void> {
  const currentUser = auth.currentUser;
  if (!currentUser) return;

  const targetId = profile.id === 'p_user_1' ? currentUser.uid : profile.id;
  if (targetId !== currentUser.uid) return;

  try {
    const userRef = doc(db, 'users', targetId);
    await setDoc(userRef, { ...profile, id: targetId, updatedAt: Date.now() }, { merge: true });
  } catch (e: any) {
    if (e?.code === 'permission-denied' || e?.message?.includes('permission')) {
      console.warn('Firestore profile sync note:', e?.message || e);
    } else {
      console.error('Error saving profile to Firestore:', e?.message || e);
    }
  }
}

/** Update User Status (online, in_match, offline) */
export async function setUserStatusInFirestore(
  uid: string, 
  status: 'online' | 'in_match' | 'offline'
): Promise<void> {
  try {
    const userRef = doc(db, 'users', uid);
    await updateDoc(userRef, { status, updatedAt: Date.now() });
  } catch (e) {
    console.error('Error updating user status:', e);
  }
}

/** Update User Stats after Match Completion */
export async function updateUserStatsInFirestore(
  uid: string, 
  matchResult: { won: boolean; mmrDelta: number; points: number; responseTimeMs: number }
): Promise<void> {
  try {
    const userRef = doc(db, 'users', uid);
    const snap = await getDoc(userRef);
    if (!snap.exists()) return;

    const current = snap.data() as PlayerProfile;
    const totalMatches = (current.totalMatches || 0) + 1;
    const wins = (current.wins || 0) + (matchResult.won ? 1 : 0);
    const losses = totalMatches - wins;
    const mmr = Math.max(100, (current.mmr || 1000) + matchResult.mmrDelta);
    const totalPoints = (current.totalPoints || 0) + matchResult.points;
    const currentStreak = matchResult.won ? (current.currentStreak || 0) + 1 : 0;
    const highestStreak = Math.max(current.highestStreak || 0, currentStreak);

    await updateDoc(userRef, {
      totalMatches,
      wins,
      losses,
      mmr,
      totalPoints,
      currentStreak,
      highestStreak,
      updatedAt: Date.now()
    });
  } catch (e) {
    console.error('Error updating user stats:', e);
  }
}

/** Real-time Leaderboard Listener derived from top MMR users */
export function listenToLeaderboard(onLeaderboardUpdated: (leaderboard: LeaderboardEntry[]) => void) {
  try {
    const topUsersQuery = query(
      collection(db, 'users'),
      orderBy('mmr', 'desc'),
      limit(50)
    );

    return onSnapshot(topUsersQuery, (snapshot) => {
      if (snapshot.empty) return;
      
      const entries: LeaderboardEntry[] = snapshot.docs.map((docSnap, index) => {
        const u = docSnap.data() as PlayerProfile;
        const total = u.totalMatches || 1;
        const wins = u.wins || 0;
        const winRate = Number(((wins / total) * 100).toFixed(1));

        let badge = '⭐ Bronze';
        if (u.mmr >= 2000) badge = '👑 Grandmaster';
        else if (u.mmr >= 1800) badge = '🥈 Master';
        else if (u.mmr >= 1500) badge = '🥉 Diamond';
        else if (u.mmr >= 1200) badge = '⭐ Gold II';

        return {
          rank: index + 1,
          id: docSnap.id,
          username: u.username || 'Anonymous',
          avatarUrl: u.avatarUrl || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
          title: u.title || '⚡ Duelist',
          mmr: u.mmr || 1000,
          wins,
          totalMatches: total,
          winRate,
          avgResponseTimeMs: u.avgResponseTimeMs || 2000,
          accuracyPercentage: u.accuracyPercentage || 75,
          badge,
        };
      });

      onLeaderboardUpdated(entries);
    }, (error) => {
      console.warn('Leaderboard listener warning:', error);
    });
  } catch (e) {
    console.error('Error listening to leaderboard:', e);
    return () => {};
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// 3. FIRESTORE TABLE: 'matches' (Real-time 1v1 Dual Exam Rooms)
// ═══════════════════════════════════════════════════════════════════════════

/** Real-time Match State Sync */
export function listenToMatch(matchId: string, onMatchUpdated: (match: MatchState) => void) {
  try {
    const matchRef = doc(db, 'matches', matchId);
    return onSnapshot(
      matchRef, 
      (snapshot) => {
        if (snapshot.exists()) {
          onMatchUpdated(snapshot.data() as MatchState);
        }
      },
      (error) => {
        console.warn('Match listener snapshot note:', error?.message || error);
      }
    );
  } catch (e) {
    console.error('Error listening to match room:', e);
    return () => {};
  }
}

/** Save or Update 1v1 Match State */
export async function saveMatchToFirestore(match: MatchState): Promise<void> {
  try {
    const matchRef = doc(db, 'matches', match.id);
    await setDoc(matchRef, { ...match, updatedAt: Date.now() }, { merge: true });
  } catch (e) {
    console.error('Error saving match to Firestore:', e);
  }
}

/** Helper to convert PlayerProfile to MatchPlayer */
export function convertProfileToMatchPlayer(profile: PlayerProfile): MatchPlayer {
  return {
    id: profile.id,
    username: profile.username,
    avatarUrl: profile.avatarUrl,
    title: profile.title || '⚡ Duelist',
    mmr: profile.mmr || 1000,
    score: 0,
    answers: [],
    isReady: true,
  };
}

/** Create Open 1v1 Match Room */
export async function createMatchInFirestore(
  category: string, 
  hostProfile: PlayerProfile
): Promise<string> {
  try {
    const matchRef = doc(collection(db, 'matches'));
    const initialMatch: Partial<MatchState> = {
      id: matchRef.id,
      roomCode: Math.floor(100000 + Math.random() * 900000).toString(),
      category,
      status: 'waiting',
      player1: convertProfileToMatchPlayer(hostProfile),
      player2: SEARCHING_MATCH_PLAYER,
      questions: [],
      currentQuestionIndex: 0,
      timeLimitSeconds: 15,
      createdAt: Date.now(),
    };
    await setDoc(matchRef, initialMatch);
    return matchRef.id;
  } catch (e) {
    console.error('Error creating match in Firestore:', e);
    throw e;
  }
}

/** Join Match Room */
export async function joinMatchInFirestore(
  matchId: string, 
  guestProfile: PlayerProfile
): Promise<void> {
  try {
    const matchRef = doc(db, 'matches', matchId);
    await updateDoc(matchRef, {
      player2: convertProfileToMatchPlayer(guestProfile),
      status: 'starting',
      updatedAt: Date.now()
    });
  } catch (e) {
    console.error('Error joining match:', e);
    throw e;
  }
}

/** Listen to Open/Waiting Matches for a Category */
export function listenToOpenMatches(
  category: string, 
  onMatchesUpdated: (matches: MatchState[]) => void
) {
  try {
    const matchesQuery = query(
      collection(db, 'matches'),
      where('status', '==', 'waiting'),
      where('category', '==', category),
      orderBy('createdAt', 'desc'),
      limit(20)
    );

    return onSnapshot(
      matchesQuery, 
      (snapshot) => {
        const openMatches = snapshot.docs.map(d => d.data() as MatchState);
        onMatchesUpdated(openMatches);
      },
      (error) => {
        console.warn('Open matches listener snapshot note:', error?.message || error);
      }
    );
  } catch (e) {
    console.error('Error listening to open matches:', e);
    return () => {};
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// 4. FIRESTORE TABLE: 'globalChat' (Live Chat Room)
// ═══════════════════════════════════════════════════════════════════════════

/** Real-time Global Chat Listener */
export function listenToGlobalChat(onChatUpdated: (messages: ChatMessage[]) => void) {
  try {
    const chatQuery = query(
      collection(db, 'globalChat'),
      orderBy('timestamp', 'asc'),
      limit(50)
    );

    return onSnapshot(chatQuery, (snapshot) => {
      const messages: ChatMessage[] = snapshot.docs.map((d) => ({
        id: d.id,
        ...d.data(),
      })) as ChatMessage[];
      onChatUpdated(messages);
    }, (error) => {
      console.warn('Global chat listener fallback:', error);
    });
  } catch (e) {
    console.error('Error listening to global chat:', e);
    return () => {};
  }
}

/** Post Chat Message */
export async function sendGlobalChatMessage(message: Omit<ChatMessage, 'id'>): Promise<void> {
  if (!auth.currentUser) return;
  try {
    await addDoc(collection(db, 'globalChat'), {
      ...message,
      senderId: auth.currentUser.uid,
      timestamp: Date.now(),
    });
  } catch (e) {
    console.error('Error sending chat message:', e);
  }
}

// ═══════════════════════════════════════════════════════════════════════════
// 5. FIRESTORE TABLE: 'notifications' (User Invites & Notifications)
// ═══════════════════════════════════════════════════════════════════════════

/** Listen to User Notifications */
export function listenToNotifications(
  recipientId: string, 
  onNotificationsUpdated: (notifications: AppNotification[]) => void
) {
  if (!recipientId) return () => {};
  try {
    const notifQuery = query(
      collection(db, 'notifications'),
      where('recipientId', '==', recipientId),
      orderBy('timestamp', 'desc'),
      limit(30)
    );

    return onSnapshot(
      notifQuery, 
      (snapshot) => {
        const notifs: AppNotification[] = snapshot.docs.map(d => ({
          id: d.id,
          ...d.data()
        })) as AppNotification[];
        onNotificationsUpdated(notifs);
      },
      (error) => {
        console.warn('Notifications listener snapshot note:', error?.message || error);
      }
    );
  } catch (e) {
    console.error('Error listening to notifications:', e);
    return () => {};
  }
}

/** Send Notification / Invite to User */
export async function sendNotificationToUser(
  recipientId: string, 
  notification: Omit<AppNotification, 'id' | 'recipientId' | 'timestamp' | 'read'>
): Promise<void> {
  try {
    await addDoc(collection(db, 'notifications'), {
      ...notification,
      recipientId,
      timestamp: Date.now(),
      read: false,
    });
  } catch (e) {
    console.error('Error sending notification:', e);
  }
}

/** Mark Notification as Read */
export async function markNotificationAsRead(notificationId: string): Promise<void> {
  try {
    const notifRef = doc(db, 'notifications', notificationId);
    await updateDoc(notifRef, { read: true });
  } catch (e) {
    console.error('Error marking notification read:', e);
  }
}
