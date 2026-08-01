
Use npm

Use a <script> tag
If you're already using npm and a module bundler such as webpack or Rollup, you can run the following command to install the latest SDK (Learn more):

npm install firebase
Then, initialize Firebase and begin using the SDKs for the products you'd like to use.

// Import the functions you need from the SDKs you need
import { initializeApp } from "firebase/app";
// TODO: Add SDKs for Firebase products that you want to use
// https://firebase.google.com/docs/web/setup#available-libraries

// Your web app's Firebase configuration
const firebaseConfig = {
  apiKey: "AIzaSyAVpqoSDB9Q6YfdsbxTmvXsBQHb5lwAnKg",
  authDomain: "studio-372167647-6d5ec.firebaseapp.com",
  projectId: "studio-372167647-6d5ec",
  storageBucket: "studio-372167647-6d5ec.firebasestorage.app",
  messagingSenderId: "682930469107",
  appId: "1:682930469107:web:fe1f8d6730caa1c3391f86"
};

// Initialize Firebase
const app = initializeApp(firebaseConfig);
Note: This option uses the modular JavaScript SDK, which provides reduced SDK size.

Learn more about Firebase for web: Get Started, Web SDK API Reference, Samples