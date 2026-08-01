import { initializeApp, getApps, getApp } from 'firebase/app';
import { 
  getAuth, 
  signInAnonymously, 
  GoogleAuthProvider, 
  signInWithPopup, 
  signOut, 
  onAuthStateChanged,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  sendPasswordResetEmail,
  updateProfile,
  User 
} from 'firebase/auth';
import { getFirestore } from 'firebase/firestore';
import firebaseConfigJson from '../firebase-applet-config.json';

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY || firebaseConfigJson.apiKey,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN || firebaseConfigJson.authDomain,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID || firebaseConfigJson.projectId,
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET || firebaseConfigJson.storageBucket,
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID || firebaseConfigJson.messagingSenderId,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID || firebaseConfigJson.appId,
};

const app = !getApps().length ? initializeApp(firebaseConfig) : getApp();

const dbId = process.env.NEXT_PUBLIC_FIREBASE_FIRESTORE_DB_ID || firebaseConfigJson.firestoreDatabaseId;
export const db = dbId && dbId !== '(default)'
  ? getFirestore(app, dbId)
  : getFirestore(app);

export const auth = getAuth(app);

export { 
  signInAnonymously, 
  GoogleAuthProvider, 
  signInWithPopup, 
  signOut, 
  onAuthStateChanged,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  sendPasswordResetEmail,
  updateProfile,
};

export type { User };
