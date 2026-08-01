import { 
  PlayerProfile, 
  Question, 
  Achievement, 
  LeaderboardEntry, 
  Friend, 
  AppNotification, 
  ChatMessage,
  QuestionCategory
} from './types';

export const INITIAL_USER_PROFILE: PlayerProfile = {
  id: 'user_player_1',
  username: 'Duelist',
  avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
  title: '⚡ Speed Scholar',
  mmr: 1000,
  totalMatches: 0,
  wins: 0,
  losses: 0,
  totalPoints: 0,
  avgResponseTimeMs: 0,
  accuracyPercentage: 0,
  highestStreak: 0,
  currentStreak: 0,
  unlockedAchievements: [],
  favoriteCategory: 'Computer Science',
  status: 'online',
  bio: 'Ready for 1v1 speed duels!',
  joinedDate: 'August 2026'
};

export const INITIAL_ACHIEVEMENTS: Achievement[] = [
  {
    id: 'ach_first_win',
    title: 'First Blood',
    description: 'Win your very first 1v1 Dual Exam match',
    icon: '⚔️',
    rarity: 'Common',
    category: 'Wins',
    progress: 1,
    maxProgress: 1,
    isUnlocked: true,
    unlockedAt: '2026-07-28'
  },
  {
    id: 'ach_speed_demon',
    title: 'Speed Demon',
    description: 'Answer a question correctly in under 1.5 seconds',
    icon: '⚡',
    rarity: 'Rare',
    category: 'Speed',
    progress: 1,
    maxProgress: 1,
    isUnlocked: true,
    unlockedAt: '2026-07-29'
  },
  {
    id: 'ach_perfect_match',
    title: 'Flawless Victory',
    description: 'Win a match with 100% accuracy',
    icon: '🎯',
    rarity: 'Epic',
    category: 'Accuracy',
    progress: 1,
    maxProgress: 1,
    isUnlocked: false
  },
  {
    id: 'ach_streak_3',
    title: 'Hot Streak',
    description: 'Achieve a 3-match win streak',
    icon: '🔥',
    rarity: 'Rare',
    category: 'Wins',
    progress: 3,
    maxProgress: 3,
    isUnlocked: true,
    unlockedAt: '2026-07-30'
  },
  {
    id: 'ach_streak_10',
    title: 'Unstoppable Titan',
    description: 'Achieve a 10-match win streak',
    icon: '👑',
    rarity: 'Legendary',
    category: 'Wins',
    progress: 3,
    maxProgress: 10,
    isUnlocked: false
  },
  {
    id: 'ach_scholar',
    title: 'Exam Veteran',
    description: 'Complete 50 dual exam matches',
    icon: '🧠',
    rarity: 'Rare',
    category: 'Mastery',
    progress: 24,
    maxProgress: 50,
    isUnlocked: false
  },
  {
    id: 'ach_polymath',
    title: 'Universal Genius',
    description: 'Win matches in 5 different subject categories',
    icon: '🌐',
    rarity: 'Epic',
    category: 'Mastery',
    progress: 3,
    maxProgress: 5,
    isUnlocked: false
  },
  {
    id: 'ach_social',
    title: 'Duelist Friend',
    description: 'Complete 5 matches against added friends',
    icon: '🤝',
    rarity: 'Common',
    category: 'Social',
    progress: 2,
    maxProgress: 5,
    isUnlocked: false
  }
];

export const MOCK_LEADERBOARD: LeaderboardEntry[] = [
  {
    rank: 1,
    id: 'p_apex',
    username: 'QuantumMind',
    avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250',
    title: '👑 Grandmaster Scholar',
    mmr: 2150,
    wins: 142,
    totalMatches: 160,
    winRate: 88.7,
    avgResponseTimeMs: 1120,
    accuracyPercentage: 94,
    badge: '🥇 Grandmaster'
  },
  {
    rank: 2,
    id: 'p_byte',
    username: 'NeuralSpeed',
    avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250',
    title: '⚡ Lightning Reflexes',
    mmr: 1980,
    wins: 115,
    totalMatches: 135,
    winRate: 85.1,
    avgResponseTimeMs: 1240,
    accuracyPercentage: 91,
    badge: '🥈 Master'
  },
  {
    rank: 3,
    id: 'p_cyber',
    username: 'CyberAura',
    avatarUrl: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&q=80&w=250',
    title: '🎓 High Honor',
    mmr: 1820,
    wins: 98,
    totalMatches: 120,
    winRate: 81.6,
    avgResponseTimeMs: 1410,
    accuracyPercentage: 89,
    badge: '🥉 Diamond'
  },
  {
    rank: 4,
    id: 'user_player_1', // Current user
    username: 'ReflexMaster',
    avatarUrl: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&q=80&w=250',
    title: '⚡ Speed Scholar',
    mmr: 1250,
    wins: 18,
    totalMatches: 24,
    winRate: 75.0,
    avgResponseTimeMs: 1680,
    accuracyPercentage: 88,
    badge: '⭐ Gold II'
  },
  {
    rank: 5,
    id: 'p_bot_dr_apex',
    username: 'Dr. Apex AI',
    avatarUrl: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&q=80&w=250',
    title: '🤖 Expert Exam Bot',
    mmr: 1210,
    wins: 45,
    totalMatches: 62,
    winRate: 72.5,
    avgResponseTimeMs: 1850,
    accuracyPercentage: 84,
    badge: '⭐ Gold I'
  }
];

export const INITIAL_FRIENDS: Friend[] = [
  {
    id: 'p_apex',
    username: 'QuantumMind',
    avatarUrl: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250',
    title: '👑 Grandmaster Scholar',
    mmr: 2150,
    status: 'online',
    winRate: 88.7,
    isFavorite: true
  },
  {
    id: 'p_byte',
    username: 'NeuralSpeed',
    avatarUrl: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=250',
    title: '⚡ Lightning Reflexes',
    mmr: 1980,
    status: 'in_match',
    winRate: 85.1,
    isFavorite: true
  },
  {
    id: 'p_cyber',
    username: 'CyberAura',
    avatarUrl: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&q=80&w=250',
    title: '🎓 High Honor',
    mmr: 1820,
    status: 'offline',
    winRate: 81.6
  },
  {
    id: 'p_sophia',
    username: 'SophiaMath',
    avatarUrl: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?auto=format&fit=crop&q=80&w=250',
    title: '🧮 Logic Queen',
    mmr: 1390,
    status: 'online',
    winRate: 71.0
  }
];

export const INITIAL_NOTIFICATIONS: AppNotification[] = [
  {
    id: 'n1',
    type: 'challenge',
    title: '1v1 Dual Challenge Invited',
    message: 'QuantumMind sent you a 1v1 challenge in Computer Science!',
    timestamp: 1775000000000,
    read: false,
    actionPayload: { roomCode: 'EXAM-99' }
  },
  {
    id: 'n2',
    type: 'achievement',
    title: 'Achievement Unlocked: Speed Demon!',
    message: 'You answered a question in 1.4 seconds with 100% accuracy.',
    timestamp: 1774900000000,
    read: true
  }
];

export const INITIAL_GLOBAL_CHAT: ChatMessage[] = [
  {
    id: 'm1',
    senderId: 'p_apex',
    senderName: 'QuantumMind',
    senderAvatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=250',
    senderTitle: '👑 Grandmaster',
    text: 'Who is ready for a 1v1 Dual Exam in Computer Science? ⚡',
    timestamp: 1775000000000
  },
  {
    id: 'm2',
    senderId: 'p_sophia',
    senderName: 'SophiaMath',
    senderAvatar: 'https://images.unsplash.com/photo-1524504388940-b1c1722653e1?auto=format&fit=crop&q=80&w=250',
    senderTitle: '🧮 Logic Queen',
    text: 'Just hit a 1.2s answer speed on calculus! Challenge me in Math lobby!',
    timestamp: 1775000300000
  }
];

export const SAMPLE_QUESTIONS: Record<QuestionCategory, Question[]> = {
  'Computer Science': [
    {
      id: 'cs1',
      category: 'Computer Science',
      question: 'Which data structure operates strictly on a First-In, First-Out (FIFO) principle?',
      options: ['Stack', 'Queue', 'Binary Search Tree', 'Priority Queue'],
      correctAnswerIndex: 1,
      explanation: 'A Queue processes elements in First-In, First-Out (FIFO) order, whereas a Stack is Last-In, First-Out (LIFO).',
      difficulty: 'Easy'
    },
    {
      id: 'cs2',
      category: 'Computer Science',
      question: 'What is the average time complexity of searching in a balanced Binary Search Tree (BST)?',
      options: ['O(1)', 'O(n)', 'O(log n)', 'O(n log n)'],
      correctAnswerIndex: 2,
      explanation: 'A balanced BST eliminates half the remaining nodes with each comparison, resulting in O(log n) time complexity.',
      difficulty: 'Medium'
    },
    {
      id: 'cs3',
      category: 'Computer Science',
      question: 'Which HTTP status code signifies that a client is Unauthorized to access a resource?',
      options: ['200 OK', '401 Unauthorized', '403 Forbidden', '404 Not Found'],
      correctAnswerIndex: 1,
      explanation: '401 indicates authentication credentials are missing or invalid. 403 means authenticated but lack permission.',
      difficulty: 'Easy'
    },
    {
      id: 'cs4',
      category: 'Computer Science',
      question: 'In garbage-collected runtimes like V8, what algorithm is primarily used for tracing unreferenced objects?',
      options: ['Mark-and-Sweep', 'Bubble Sort', 'Reference Counting', 'Dijkstra Algorithm'],
      correctAnswerIndex: 0,
      explanation: 'Mark-and-Sweep traverses root objects, marks reachable objects, and sweeps away unmarked allocations.',
      difficulty: 'Hard'
    },
    {
      id: 'cs5',
      category: 'Computer Science',
      question: 'What does the ACID acronym stand for in relational database management systems?',
      options: [
        'Atomicity, Consistency, Isolation, Durability',
        'Asynchronous, Concurrent, Isolated, Distributed',
        'Allocation, Cache, Indexing, Data',
        'API, Control, Interface, Dependency'
      ],
      correctAnswerIndex: 0,
      explanation: 'ACID guarantees database transaction reliability through Atomicity, Consistency, Isolation, and Durability.',
      difficulty: 'Medium'
    }
  ],
  'Science & Physics': [
    {
      id: 'sci1',
      category: 'Science & Physics',
      question: 'What is the speed of light in a vacuum (approximately)?',
      options: ['300,000 km/s', '150,000 km/s', '1,000,000 km/s', '30,000 km/s'],
      correctAnswerIndex: 0,
      explanation: 'The speed of light in vacuum (c) is precisely 299,792,458 m/s (~300,000 km/s).',
      difficulty: 'Easy'
    },
    {
      id: 'sci2',
      category: 'Science & Physics',
      question: 'Which subatomic particle carries a negative electric charge?',
      options: ['Proton', 'Neutron', 'Electron', 'Positron'],
      correctAnswerIndex: 2,
      explanation: 'Electrons carry a negative fundamental charge (-1e), while protons are positive and neutrons neutral.',
      difficulty: 'Easy'
    },
    {
      id: 'sci3',
      category: 'Science & Physics',
      question: 'What is Newton’s Second Law of Motion expressed as a mathematical formula?',
      options: ['E = mc²', 'F = ma', 'P = IV', 'V = IR'],
      correctAnswerIndex: 1,
      explanation: 'Force equals mass times acceleration (F = ma).',
      difficulty: 'Easy'
    },
    {
      id: 'sci4',
      category: 'Science & Physics',
      question: 'In thermodynamics, what physical property does the Third Law state approaches a constant value at absolute zero?',
      options: ['Enthalpy', 'Entropy', 'Temperature', 'Pressure'],
      correctAnswerIndex: 1,
      explanation: 'The Third Law states that the entropy of a perfect crystal approaches zero as temperature approaches absolute zero.',
      difficulty: 'Hard'
    },
    {
      id: 'sci5',
      category: 'Science & Physics',
      question: 'What is the primary gas that composes the atmosphere of Earth?',
      options: ['Oxygen', 'Carbon Dioxide', 'Nitrogen', 'Argon'],
      correctAnswerIndex: 2,
      explanation: 'Nitrogen makes up approximately 78% of Earth atmosphere, followed by Oxygen at 21%.',
      difficulty: 'Easy'
    }
  ],
  'Mathematics': [
    {
      id: 'math_m1',
      category: 'Mathematics',
      question: 'If the roots of the quadratic equation x² - 8x + k = 0 are equal, what is the value of k?',
      options: ['12', '16', '24', '64'],
      correctAnswerIndex: 1,
      explanation: 'For equal real roots, the discriminant Δ = b² - 4ac must equal zero.',
      difficulty: 'Medium',
      learningConcept: 'Quadratic Equations & Discriminants',
      formulaOrRule: 'Discriminant Δ = b² - 4ac = 0 (for equal roots)',
      stepByStepAnalysis: [
        'Step 1: Identify equation coefficients: a = 1, b = -8, c = k.',
        'Step 2: Write discriminant formula: Δ = (-8)² - 4(1)(k) = 64 - 4k.',
        'Step 3: Set Δ = 0 for equal roots: 64 - 4k = 0 => 4k = 64 => k = 16.'
      ]
    },
    {
      id: 'math_m2',
      category: 'Mathematics',
      question: 'What is the area of a right-angled triangle whose hypotenuse is 10 cm and one side is 6 cm?',
      options: ['18 cm²', '24 cm²', '30 cm²', '48 cm²'],
      correctAnswerIndex: 1,
      explanation: 'First calculate missing leg = √(10² - 6²) = 8 cm. Then Area = 1/2 × 6 × 8 = 24 cm².',
      difficulty: 'Easy',
      learningConcept: 'Pythagorean Theorem & Triangle Area',
      formulaOrRule: 'a² + b² = c² and Area = 1/2 × base × height',
      stepByStepAnalysis: [
        'Step 1: Apply Pythagorean theorem: b = √(10² - 6²) = √(100 - 36) = √64 = 8 cm.',
        'Step 2: Base = 6 cm, Height = 8 cm.',
        'Step 3: Calculate Area = 1/2 × 6 × 8 = 24 cm².'
      ]
    },
    {
      id: 'math_m3',
      category: 'Mathematics',
      question: 'Evaluate the indefinite integral: ∫ (2x + 3) dx.',
      options: ['x² + 3x + C', '2x² + 3x + C', 'x² + C', '2x + C'],
      correctAnswerIndex: 0,
      explanation: 'Integrate term-by-term using the power rule: ∫ 2x dx = x², ∫ 3 dx = 3x.',
      difficulty: 'Medium',
      learningConcept: 'Calculus - Indefinite Integration',
      formulaOrRule: '∫ xⁿ dx = (xⁿ⁺¹ / (n+1)) + C',
      stepByStepAnalysis: [
        'Step 1: Split integral: ∫ 2x dx + ∫ 3 dx.',
        'Step 2: Integrate 2x: 2 × (x²/2) = x².',
        'Step 3: Integrate constant 3: 3x.',
        'Step 4: Combine with constant C: x² + 3x + C.'
      ]
    },
    {
      id: 'math_m4',
      category: 'Mathematics',
      question: 'A shopkeeper marks an item 25% above cost price and then gives a 10% discount. What is the net profit %?',
      options: ['12.5%', '15%', '17.5%', '20%'],
      correctAnswerIndex: 0,
      explanation: 'Assuming Cost Price = $100, Marked Price = $125. Selling Price = $125 - 10%($125) = $112.50. Profit = 12.5%.',
      difficulty: 'Medium',
      learningConcept: 'Percentage & Profit-Loss Economics',
      formulaOrRule: 'Effective Profit % = [(Selling Price - Cost Price) / Cost Price] × 100',
      stepByStepAnalysis: [
        'Step 1: Assume Cost Price (CP) = $100.',
        'Step 2: Marked Price (MP) = $100 + 25% = $125.',
        'Step 3: Selling Price (SP) = $125 - 10%($125) = $125 - $12.50 = $112.50.',
        'Step 4: Profit = SP - CP = $112.50 - $100 = $12.50 (12.5%).'
      ]
    },
    {
      id: 'math_m5',
      category: 'Mathematics',
      question: 'If log₁₀(x) + log₁₀(x - 3) = 1, what is the valid value of x?',
      options: ['x = 2', 'x = 5', 'x = 10', 'x = -2'],
      correctAnswerIndex: 1,
      explanation: 'Combine logs: log₁₀(x² - 3x) = 1 => x² - 3x - 10 = 0. Roots are x = 5 and x = -2 (discard negative).',
      difficulty: 'Hard',
      learningConcept: 'Logarithmic Rules & Domain Constraints',
      formulaOrRule: 'log(a) + log(b) = log(a × b) and log_b(y) = c => y = b^c',
      stepByStepAnalysis: [
        'Step 1: Apply log product property: log₁₀(x(x - 3)) = 1.',
        'Step 2: Convert to exponential form: x(x - 3) = 10¹ => x² - 3x - 10 = 0.',
        'Step 3: Factor: (x - 5)(x + 2) = 0.',
        'Step 4: Log domain requires x > 0, so discard x = -2. Valid root x = 5.'
      ]
    }
  ],

  'English Language': [
    {
      id: 'eng_e1',
      category: 'English Language',
      question: 'Identify the grammatically correct sentence from the choices below:',
      options: [
        'Neither the teacher nor the students was present in class.',
        'Neither the teacher nor the students were present in class.',
        'Neither the teacher or the students was present in class.',
        'Neither the teacher nor the students has present in class.'
      ],
      correctAnswerIndex: 1,
      explanation: 'When subjects are connected by "Neither... nor", the verb agrees with the closer subject ("students" - plural).',
      difficulty: 'Medium',
      learningConcept: 'Subject-Verb Agreement (Proximity Rule)',
      formulaOrRule: 'Neither [Subject 1] nor [Subject 2] + Verb agreeing with Subject 2',
      stepByStepAnalysis: [
        'Step 1: Identify subjects: "teacher" (singular) and "students" (plural).',
        'Step 2: Identify closer subject to the verb: "students" (plural).',
        'Step 3: Match plural subject "students" with plural verb "were".'
      ]
    },
    {
      id: 'eng_e2',
      category: 'English Language',
      question: 'Choose the option that best expresses the meaning of the idiom: "To bite the bullet"',
      options: [
        'To accept a painful or difficult situation bravely',
        'To act impulsively without considering consequences',
        'To avoid taking responsibility for a mistake',
        'To start an argument unnecessarily'
      ],
      correctAnswerIndex: 0,
      explanation: '"To bite the bullet" means enduring a painful or difficult situation courageously.',
      difficulty: 'Easy',
      learningConcept: 'Idioms & Phrase Meaning',
      formulaOrRule: 'Etymology: Biting lead bullets during battlefield surgery without anesthesia',
      stepByStepAnalysis: [
        'Step 1: Deconstruct figurative imagery: Biting down to withstand extreme pressure or pain.',
        'Step 2: Translate to modern usage: Facing unavoidable hardship with stoic fortitude.',
        'Step 3: Select "To accept a painful or difficult situation bravely".'
      ]
    },
    {
      id: 'eng_e3',
      category: 'English Language',
      question: 'Select the word that is most nearly OPPOSITE in meaning to "GARRULOUS":',
      options: ['Taciturn', 'Eloquent', 'Voluble', 'Loquacious'],
      correctAnswerIndex: 0,
      explanation: 'Garrulous means excessively talkative. Taciturn means quiet and reserved.',
      difficulty: 'Medium',
      learningConcept: 'Antonyms & Advanced Vocabulary',
      formulaOrRule: 'Garrulous = Talkative; Taciturn = Uncommunicative',
      stepByStepAnalysis: [
        'Step 1: Define "Garrulous": Excessively talkative on trivial matters.',
        'Step 2: Analyze options: Voluble and Loquacious are synonyms.',
        'Step 3: Identify "Taciturn" as the exact antonym (habitually silent).'
      ]
    },
    {
      id: 'eng_e4',
      category: 'English Language',
      question: 'Which of the following sentences correctly utilizes the subjunctive mood?',
      options: [
        'If I was you, I would study harder.',
        'If I were you, I would study harder.',
        'If I am you, I will study harder.',
        'If I have been you, I had studied harder.'
      ],
      correctAnswerIndex: 1,
      explanation: 'In subjunctive hypotheticals, "were" is used for all persons regardless of singular/plural subject.',
      difficulty: 'Hard',
      learningConcept: 'Grammar - Subjunctive Mood in Hypotheticals',
      formulaOrRule: 'If + Subject + WERE (unreal condition), Subject + WOULD + Verb',
      stepByStepAnalysis: [
        'Step 1: Identify hypothetical condition (impossible reality: "I" being "you").',
        'Step 2: Apply subjunctive rule: "was" changes to "were" in counterfactual conditional clauses.',
        'Step 3: Confirm sentence: "If I were you, I would study harder."'
      ]
    },
    {
      id: 'eng_e5',
      category: 'English Language',
      question: 'Choose the correctly punctuated sentence introducing a list after a quotation:',
      options: [
        'The manager said, "We need three things: dedication, focus, and teamwork."',
        'The manager said "We need three things, dedication, focus, and teamwork."',
        'The manager said; "We need three things: dedication focus and teamwork."',
        'The manager said, "We need three things; dedication, focus and teamwork."'
      ],
      correctAnswerIndex: 0,
      explanation: 'Commas introduce quotes, colons introduce lists, and commas separate listed items.',
      difficulty: 'Easy',
      learningConcept: 'Punctuation & Direct Speech Formatting',
      formulaOrRule: 'Reporting verb + comma + quote + colon before list items',
      stepByStepAnalysis: [
        'Step 1: Reporting clause requires comma before opening quote: said, "...".',
        'Step 2: Full clause introducing list requires colon (:): "We need three things: ...".',
        'Step 3: Items in series separated by commas: dedication, focus, and teamwork.'
      ]
    }
  ],

  'Logical Reasoning': [
    {
      id: 'lr_l1',
      category: 'Logical Reasoning',
      question: 'Statements: (1) All cats are felines. (2) All felines are carnivores. What conclusion must follow?',
      options: [
        'All cats are carnivores',
        'All carnivores are cats',
        'Some cats are not felines',
        'No felines are carnivores'
      ],
      correctAnswerIndex: 0,
      explanation: 'Transitive property of categorical sets: Cats ⊂ Felines ⊂ Carnivores => All cats are carnivores.',
      difficulty: 'Easy',
      learningConcept: 'Categorical Syllogisms & Venn Inclusions',
      formulaOrRule: 'If A ⊆ B and B ⊆ C, then A ⊆ C',
      stepByStepAnalysis: [
        'Step 1: Diagram set 1: Cats circle is completely inside Felines circle.',
        'Step 2: Diagram set 2: Felines circle is completely inside Carnivores circle.',
        'Step 3: Deduce transitive overlap: Cats are entirely within Carnivores.'
      ]
    },
    {
      id: 'lr_l2',
      category: 'Logical Reasoning',
      question: 'Complete the logical number series: 2, 6, 12, 20, 30, ?',
      options: ['38', '40', '42', '44'],
      correctAnswerIndex: 2,
      explanation: 'Differences are +4, +6, +8, +10, +12. 30 + 12 = 42. (Or n(n+1) pattern: 1×2, 2×3, 3×4, 4×5, 5×6, 6×7=42).',
      difficulty: 'Medium',
      learningConcept: 'Number Series & Pattern Extraction',
      formulaOrRule: 'Difference series (+2 increment) or Formula T_n = n × (n + 1)',
      stepByStepAnalysis: [
        'Step 1: Calculate consecutive differences: 6-2=4, 12-6=6, 20-12=8, 30-20=10.',
        'Step 2: Identify difference progression: +4, +6, +8, +10 (increases by 2 each step).',
        'Step 3: Next difference is +12. Add to 30: 30 + 12 = 42.'
      ]
    },
    {
      id: 'lr_l3',
      category: 'Logical Reasoning',
      question: 'In a code language, "BRAIN" is coded as "CSBJO". How is "STORM" coded in that language?',
      options: ['TUPSN', 'TURSN', 'TVPSN', 'TUPSM'],
      correctAnswerIndex: 0,
      explanation: 'Each letter is shifted forward by +1 position in the alphabet: S->T, T->U, O->P, R->S, M->N.',
      difficulty: 'Easy',
      learningConcept: 'Alphabetical Shift & Cipher Decoding',
      formulaOrRule: 'Cipher Rule: Letter(n) -> Letter(n + 1)',
      stepByStepAnalysis: [
        'Step 1: Test pattern on BRAIN: B(+1)=C, R(+1)=S, A(+1)=B, I(+1)=J, N(+1)=O.',
        'Step 2: Apply +1 shift to STORM: S+1=T, T+1=U, O+1=P, R+1=S, M+1=N.',
        'Step 3: Combine letters: TUPSN.'
      ]
    },
    {
      id: 'lr_l4',
      category: 'Logical Reasoning',
      question: 'Rahul said: "She is the daughter of my grandfather\'s only son." How is the girl related to Rahul?',
      options: ['Mother', 'Sister', 'Aunt', 'Cousin'],
      correctAnswerIndex: 1,
      explanation: 'Grandfather\'s only son = Rahul\'s father. Daughter of Rahul\'s father = Rahul\'s sister.',
      difficulty: 'Medium',
      learningConcept: 'Blood Relations & Family Tree Analysis',
      formulaOrRule: 'Grandfather\'s only son = Father; Father\'s daughter = Sister',
      stepByStepAnalysis: [
        'Step 1: Decode "my grandfather\'s only son": That refers to Rahul\'s father.',
        'Step 2: Decode "daughter of my father": That refers to Rahul\'s sister.',
        'Step 3: Conclude relationship: The girl is Rahul\'s sister.'
      ]
    },
    {
      id: 'lr_l5',
      category: 'Logical Reasoning',
      question: 'Five friends A, B, C, D, E sit in a row facing North. A sits next to B. C sits next to D. D is not next to E (who is at far left). C is 2nd from right. Who is in the middle?',
      options: ['A', 'B', 'C', 'D'],
      correctAnswerIndex: 3,
      explanation: 'Positions 1 to 5: E is at 1. C is at 4. D sits next to C and not E, so D is at 3 (middle).',
      difficulty: 'Hard',
      learningConcept: 'Linear Arrangement & Constraint Elimination',
      formulaOrRule: 'Place fixed anchors first (Pos 1 & Pos 4), then enforce non-adjacency rules',
      stepByStepAnalysis: [
        'Step 1: Set position anchors: E is at Pos 1 (far left). C is at Pos 4 (2nd from right).',
        'Step 2: C sits next to D. D could be Pos 3 or 5. But D cannot sit near E.',
        'Step 3: Place D at Pos 3 (middle). The row becomes E _ D C _.',
        'Step 4: Middle position (Pos 3) is uniquely occupied by D.'
      ]
    }
  ],

  'Mathematics & Logic': [
    {
      id: 'math1',
      category: 'Mathematics & Logic',
      question: 'What is the value of 15% of 240?',
      options: ['30', '36', '42', '24'],
      correctAnswerIndex: 1,
      explanation: '10% of 240 is 24, and 5% is 12. 24 + 12 = 36.',
      difficulty: 'Easy',
      learningConcept: 'Fast Mental Percentages',
      formulaOrRule: 'Percentage = (Value / 100) × Base',
      stepByStepAnalysis: [
        'Step 1: Find 10% of 240 by shifting decimal point = 24.',
        'Step 2: Find 5% by taking half of 10% = 12.',
        'Step 3: Add 10% + 5% = 24 + 12 = 36.'
      ]
    },
    {
      id: 'math2',
      category: 'Mathematics & Logic',
      question: 'Solve for x: 3x + 12 = 33',
      options: ['x = 5', 'x = 7', 'x = 9', 'x = 11'],
      correctAnswerIndex: 1,
      explanation: 'Subtract 12 from both sides: 3x = 21. Divide by 3: x = 7.',
      difficulty: 'Easy',
      learningConcept: 'Linear Algebraic Equations',
      formulaOrRule: 'Isolate variable term by inverse arithmetic operations',
      stepByStepAnalysis: [
        'Step 1: Subtract 12 from both sides: 3x = 33 - 12 = 21.',
        'Step 2: Divide both sides by 3: x = 21 / 3 = 7.'
      ]
    },
    {
      id: 'math3',
      category: 'Mathematics & Logic',
      question: 'What is the derivative of f(x) = x³ with respect to x?',
      options: ['3x²', 'x²', '3x', '6x'],
      correctAnswerIndex: 0,
      explanation: 'Using the power rule d/dx(x^n) = n*x^(n-1), d/dx(x³) = 3x².',
      difficulty: 'Medium',
      learningConcept: 'Calculus Differentiation',
      formulaOrRule: 'Power Rule: d/dx(xⁿ) = n·xⁿ⁻¹',
      stepByStepAnalysis: [
        'Step 1: Identify exponent n = 3.',
        'Step 2: Multiply function by exponent: 3 · x.',
        'Step 3: Subtract 1 from exponent: 3 - 1 = 2 => Result: 3x².'
      ]
    },
    {
      id: 'math4',
      category: 'Mathematics & Logic',
      question: 'If a coin is flipped 3 times, what is the probability of getting exactly 2 heads?',
      options: ['1/8', '3/8', '1/2', '5/8'],
      correctAnswerIndex: 1,
      explanation: 'Total outcomes = 2³ = 8. Favorable outcomes (HHT, HTH, THH) = 3. Probability = 3/8.',
      difficulty: 'Medium',
      learningConcept: 'Combinatorial Probability',
      formulaOrRule: 'P(Event) = Favorable Outcomes / Total Outcomes',
      stepByStepAnalysis: [
        'Step 1: Calculate total possible outcomes = 2 × 2 × 2 = 8.',
        'Step 2: List outcomes with exactly 2 Heads: {HHT, HTH, THH} = 3 outcomes.',
        'Step 3: Compute probability = 3 / 8.'
      ]
    },
    {
      id: 'math5',
      category: 'Mathematics & Logic',
      question: 'What is the sum of interior angles of a regular hexagon?',
      options: ['540°', '720°', '900°', '1080°'],
      correctAnswerIndex: 1,
      explanation: 'Formula for interior angle sum: (n - 2) * 180°. For a hexagon (n=6): (6 - 2) * 180 = 720°.',
      difficulty: 'Medium',
      learningConcept: 'Polygon Geometry',
      formulaOrRule: 'Sum of interior angles = (n - 2) × 180°',
      stepByStepAnalysis: [
        'Step 1: Identify number of sides n = 6 for a hexagon.',
        'Step 2: Apply formula: (6 - 2) × 180° = 4 × 180°.',
        'Step 3: Calculate 4 × 180 = 720°.'
      ]
    }
  ],
  'General Knowledge': [
    {
      id: 'gk1',
      category: 'General Knowledge',
      question: 'Which planet in our solar system is known as the "Red Planet"?',
      options: ['Venus', 'Mars', 'Jupiter', 'Saturn'],
      correctAnswerIndex: 1,
      explanation: 'Mars is called the Red Planet due to the reddish iron oxide prevalent on its surface.',
      difficulty: 'Easy'
    },
    {
      id: 'gk2',
      category: 'General Knowledge',
      question: 'Which chemical element has the atomic symbol "Au"?',
      options: ['Silver', 'Gold', 'Copper', 'Aluminum'],
      correctAnswerIndex: 1,
      explanation: 'Au comes from the Latin word "Aurum", meaning gold.',
      difficulty: 'Easy'
    },
    {
      id: 'gk3',
      category: 'General Knowledge',
      question: 'Who painted the Mona Lisa?',
      options: ['Vincent van Gogh', 'Pablo Picasso', 'Leonardo da Vinci', 'Michelangelo'],
      correctAnswerIndex: 2,
      explanation: 'The Mona Lisa was painted by the Italian Renaissance polymath Leonardo da Vinci in the early 16th century.',
      difficulty: 'Easy'
    },
    {
      id: 'gk4',
      category: 'General Knowledge',
      question: 'What is the capital city of Australia?',
      options: ['Sydney', 'Melbourne', 'Canberra', 'Brisbane'],
      correctAnswerIndex: 2,
      explanation: 'Canberra was chosen as the capital of Australia in 1908 as a compromise between Sydney and Melbourne.',
      difficulty: 'Medium'
    },
    {
      id: 'gk5',
      category: 'General Knowledge',
      question: 'Which ocean is the largest and deepest on Earth?',
      options: ['Atlantic Ocean', 'Indian Ocean', 'Pacific Ocean', 'Arctic Ocean'],
      correctAnswerIndex: 2,
      explanation: 'The Pacific Ocean covers over 30% of the Earth’s surface and contains the Mariana Trench.',
      difficulty: 'Easy'
    }
  ],
  'World History': [
    {
      id: 'hist1',
      category: 'World History',
      question: 'In which year did World War II officially end?',
      options: ['1918', '1939', '1945', '1950'],
      correctAnswerIndex: 2,
      explanation: 'World War II ended in 1945 following the surrender of Germany in May and Japan in September.',
      difficulty: 'Easy'
    },
    {
      id: 'hist2',
      category: 'World History',
      question: 'Who was the first President of the United States?',
      options: ['Thomas Jefferson', 'George Washington', 'Benjamin Franklin', 'John Adams'],
      correctAnswerIndex: 1,
      explanation: 'George Washington served as the 1st President of the United States from 1789 to 1797.',
      difficulty: 'Easy'
    },
    {
      id: 'hist3',
      category: 'World History',
      question: 'Which ancient civilization built the Pyramids of Giza?',
      options: ['Mesopotamians', 'Ancient Egyptians', 'Romans', 'Greeks'],
      correctAnswerIndex: 1,
      explanation: 'Ancient Egyptians built the Pyramids of Giza during the Old Kingdom period.',
      difficulty: 'Easy'
    },
    {
      id: 'hist4',
      category: 'World History',
      question: 'The Magna Carta was signed by King John of England in which year?',
      options: ['1066', '1215', '1492', '1776'],
      correctAnswerIndex: 1,
      explanation: 'The Magna Carta was agreed to by King John at Runnymede in 1215.',
      difficulty: 'Medium'
    },
    {
      id: 'hist5',
      category: 'World History',
      question: 'Which empire was ruled by Julius Caesar and Augustus?',
      options: ['Ottoman Empire', 'Mongol Empire', 'Roman Empire', 'Byzantine Empire'],
      correctAnswerIndex: 2,
      explanation: 'Julius Caesar led the late Roman Republic and Augustus became the first Emperor of the Roman Empire.',
      difficulty: 'Easy'
    }
  ],
  'Medical & Biology': [
    {
      id: 'bio1',
      category: 'Medical & Biology',
      question: 'Which organelle is widely known as the "powerhouse of the cell"?',
      options: ['Nucleus', 'Mitochondria', 'Ribosome', 'Golgi Apparatus'],
      correctAnswerIndex: 1,
      explanation: 'Mitochondria generate most of the chemical energy needed to power cellular reactions via ATP synthesis.',
      difficulty: 'Easy'
    },
    {
      id: 'bio2',
      category: 'Medical & Biology',
      question: 'What is the normal human body temperature in Celsius?',
      options: ['35.0°C', '37.0°C', '39.0°C', '40.0°C'],
      correctAnswerIndex: 1,
      explanation: 'Average normal human body temperature is approximately 37.0°C (98.6°F).',
      difficulty: 'Easy'
    },
    {
      id: 'bio3',
      category: 'Medical & Biology',
      question: 'Which protein in red blood cells is responsible for carrying oxygen throughout the body?',
      options: ['Hemoglobin', 'Insulin', 'Collagen', 'Myosin'],
      correctAnswerIndex: 0,
      explanation: 'Hemoglobin binds oxygen in lungs and delivers it to tissues throughout the body.',
      difficulty: 'Easy'
    },
    {
      id: 'bio4',
      category: 'Medical & Biology',
      question: 'How many chambers does a healthy human heart contain?',
      options: ['2', '3', '4', '6'],
      correctAnswerIndex: 2,
      explanation: 'The human heart has 4 chambers: two upper atria and two lower ventricles.',
      difficulty: 'Easy'
    },
    {
      id: 'bio5',
      category: 'Medical & Biology',
      question: 'Which type of RNA delivers amino acids to the ribosome during protein synthesis?',
      options: ['mRNA (Messenger)', 'tRNA (Transfer)', 'rRNA (Ribosomal)', 'snRNA (Small nuclear)'],
      correctAnswerIndex: 1,
      explanation: 'tRNA carries specific amino acids corresponding to mRNA codons to assemble polypeptide chains.',
      difficulty: 'Medium'
    }
  ],
  'GRE / SAT Prep': [
    {
      id: 'sat1',
      category: 'GRE / SAT Prep',
      question: 'Choose the word that is most nearly SYNONYMOUS with "ephemeral":',
      options: ['Permanent', 'Transient', 'Substantial', 'Luminous'],
      correctAnswerIndex: 1,
      explanation: 'Ephemeral means lasting for a very short time; transient is its closest synonym.',
      difficulty: 'Medium'
    },
    {
      id: 'sat2',
      category: 'GRE / SAT Prep',
      question: 'If f(x) = 2x² - 5x + 3, what is f(-2)?',
      options: ['5', '11', '21', '25'],
      correctAnswerIndex: 2,
      explanation: 'f(-2) = 2(-2)² - 5(-2) + 3 = 2(4) + 10 + 3 = 8 + 10 + 3 = 21.',
      difficulty: 'Medium'
    },
    {
      id: 'sat3',
      category: 'GRE / SAT Prep',
      question: 'Identify the ANTONYM for the word "pragmatic":',
      options: ['Idealistic', 'Practical', 'Realistic', 'Sensible'],
      correctAnswerIndex: 0,
      explanation: 'Pragmatic means dealing with things sensibly and realistically; idealistic is its antonym.',
      difficulty: 'Medium'
    },
    {
      id: 'sat4',
      category: 'GRE / SAT Prep',
      question: 'A circle has an area of 36π. What is its circumference?',
      options: ['6π', '12π', '18π', '36π'],
      correctAnswerIndex: 1,
      explanation: 'Area = πr² = 36π => r = 6. Circumference = 2πr = 2π(6) = 12π.',
      difficulty: 'Medium'
    },
    {
      id: 'sat5',
      category: 'GRE / SAT Prep',
      question: 'Choose the correct word: "The candidate spoke with such _____ that the entire room was persuaded."',
      options: ['Eloquence', 'Apathy', 'Belligerence', 'Ambiguity'],
      correctAnswerIndex: 0,
      explanation: 'Eloquence means fluent or persuasive speaking or writing.',
      difficulty: 'Medium'
    }
  ],
  'Custom AI Topic': []
};

// Local Storage Helper
export function getStoredProfile(): PlayerProfile {
  if (typeof window === 'undefined') return INITIAL_USER_PROFILE;
  try {
    const data = localStorage.getItem('dual_exam_user_profile');
    return data ? JSON.parse(data) : INITIAL_USER_PROFILE;
  } catch (e) {
    return INITIAL_USER_PROFILE;
  }
}

export function saveStoredProfile(profile: PlayerProfile): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem('dual_exam_user_profile', JSON.stringify(profile));
  } catch (e) {
    console.error(e);
  }
}

export function getStoredAchievements(): Achievement[] {
  if (typeof window === 'undefined') return INITIAL_ACHIEVEMENTS;
  try {
    const data = localStorage.getItem('dual_exam_achievements');
    return data ? JSON.parse(data) : INITIAL_ACHIEVEMENTS;
  } catch (e) {
    return INITIAL_ACHIEVEMENTS;
  }
}

export function saveStoredAchievements(achievements: Achievement[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem('dual_exam_achievements', JSON.stringify(achievements));
  } catch (e) {
    console.error(e);
  }
}

export function getStoredFriends(): Friend[] {
  if (typeof window === 'undefined') return INITIAL_FRIENDS;
  try {
    const data = localStorage.getItem('dual_exam_friends');
    return data ? JSON.parse(data) : INITIAL_FRIENDS;
  } catch (e) {
    return INITIAL_FRIENDS;
  }
}

export function saveStoredFriends(friends: Friend[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem('dual_exam_friends', JSON.stringify(friends));
  } catch (e) {
    console.error(e);
  }
}

export function getStoredNotifications(): AppNotification[] {
  if (typeof window === 'undefined') return INITIAL_NOTIFICATIONS;
  try {
    const data = localStorage.getItem('dual_exam_notifications');
    return data ? JSON.parse(data) : INITIAL_NOTIFICATIONS;
  } catch (e) {
    return INITIAL_NOTIFICATIONS;
  }
}

export function saveStoredNotifications(notifs: AppNotification[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem('dual_exam_notifications', JSON.stringify(notifs));
  } catch (e) {
    console.error(e);
  }
}

export function getStoredGlobalChat(): ChatMessage[] {
  if (typeof window === 'undefined') return INITIAL_GLOBAL_CHAT;
  try {
    const data = localStorage.getItem('dual_exam_global_chat');
    return data ? JSON.parse(data) : INITIAL_GLOBAL_CHAT;
  } catch (e) {
    return INITIAL_GLOBAL_CHAT;
  }
}

export function saveStoredGlobalChat(messages: ChatMessage[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem('dual_exam_global_chat', JSON.stringify(messages));
  } catch (e) {
    console.error(e);
  }
}

// BroadcastChannel for cross-tab multi-window 1v1 play!
export class ArenaBroadcastEngine {
  private channel: BroadcastChannel | null = null;
  private listeners: Array<(event: MessageEvent) => void> = [];

  constructor() {
    if (typeof window !== 'undefined' && 'BroadcastChannel' in window) {
      this.channel = new BroadcastChannel('dual_exam_arena');
      this.channel.onmessage = (e) => {
        this.listeners.forEach((fn) => fn(e));
      };
    }
  }

  public publish(type: string, payload: unknown) {
    if (this.channel) {
      this.channel.postMessage({ type, payload, timestamp: Date.now() });
    }
  }

  public subscribe(fn: (event: MessageEvent) => void) {
    this.listeners.push(fn);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== fn);
    };
  }
}

export const arenaBroadcast = new ArenaBroadcastEngine();

// ── JWT Token Persistence ─────────────────────────────────────────────────
// Stores the Go backend JWT so users stay logged-in across page refreshes.

const JWT_STORAGE_KEY = 'ea_jwt';

export function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(JWT_STORAGE_KEY);
}

export function saveAuthToken(token: string): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(JWT_STORAGE_KEY, token);
}

export function clearAuthToken(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(JWT_STORAGE_KEY);
}
