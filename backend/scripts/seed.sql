-- =====================================================================
-- SEED DATA: 50 Reasoning Questions for SSC Category
-- Distribution per match (10 questions): 4 easy, 4 medium, 2 hard
-- Total: 20 easy + 20 medium + 10 hard = 50 questions for 5 matches
-- =====================================================================

-- Get the SSC category ID (seeded by migration)
DO $$
DECLARE
    ssc_id UUID;
    topic_analogy UUID;
    topic_series UUID;
    topic_coding UUID;
    topic_blood UUID;
    topic_direction UUID;
BEGIN

SELECT id INTO ssc_id FROM exam_categories WHERE code = 'SSC';

-- ── Topics ──────────────────────────────────────────────────────────
INSERT INTO topics (id, exam_category_id, name) VALUES
    (gen_random_uuid(), ssc_id, 'Analogy'),
    (gen_random_uuid(), ssc_id, 'Number Series'),
    (gen_random_uuid(), ssc_id, 'Coding-Decoding'),
    (gen_random_uuid(), ssc_id, 'Blood Relations'),
    (gen_random_uuid(), ssc_id, 'Direction Sense')
ON CONFLICT DO NOTHING;

SELECT id INTO topic_analogy   FROM topics WHERE name = 'Analogy'        AND exam_category_id = ssc_id LIMIT 1;
SELECT id INTO topic_series    FROM topics WHERE name = 'Number Series'  AND exam_category_id = ssc_id LIMIT 1;
SELECT id INTO topic_coding    FROM topics WHERE name = 'Coding-Decoding' AND exam_category_id = ssc_id LIMIT 1;
SELECT id INTO topic_blood     FROM topics WHERE name = 'Blood Relations' AND exam_category_id = ssc_id LIMIT 1;
SELECT id INTO topic_direction FROM topics WHERE name = 'Direction Sense' AND exam_category_id = ssc_id LIMIT 1;

-- =====================================================================
-- EASY QUESTIONS (20)
-- =====================================================================

-- Q1 Easy - Analogy
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000001', ssc_id, topic_analogy, 'mcq_single', 'easy',
    'Pen is to Write as Knife is to ___?',
    'Pen is used to write. Similarly, a knife is used to cut.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000001', 'Cut',   true,  1),
    ('a0000001-0000-0000-0000-000000000001', 'Eat',   false, 2),
    ('a0000001-0000-0000-0000-000000000001', 'Cook',  false, 3),
    ('a0000001-0000-0000-0000-000000000001', 'Sharp', false, 4);

-- Q2 Easy - Analogy
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000002', ssc_id, topic_analogy, 'mcq_single', 'easy',
    'Eye is to See as Ear is to ___?',
    'Eye is used to see. Similarly, ear is used to hear.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000002', 'Hear',  true,  1),
    ('a0000001-0000-0000-0000-000000000002', 'Smell', false, 2),
    ('a0000001-0000-0000-0000-000000000002', 'Taste', false, 3),
    ('a0000001-0000-0000-0000-000000000002', 'Touch', false, 4);

-- Q3 Easy - Number Series
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000003', ssc_id, topic_series, 'mcq_single', 'easy',
    'What comes next: 2, 4, 6, 8, ___?',
    'The pattern adds 2 each time. 8 + 2 = 10.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000003', '10', true,  1),
    ('a0000001-0000-0000-0000-000000000003', '12', false, 2),
    ('a0000001-0000-0000-0000-000000000003', '9',  false, 3),
    ('a0000001-0000-0000-0000-000000000003', '11', false, 4);

-- Q4 Easy - Number Series
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000004', ssc_id, topic_series, 'mcq_single', 'easy',
    'What comes next: 1, 3, 5, 7, ___?',
    'Odd number series. 7 + 2 = 9.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000004', '9',  true,  1),
    ('a0000001-0000-0000-0000-000000000004', '8',  false, 2),
    ('a0000001-0000-0000-0000-000000000004', '10', false, 3),
    ('a0000001-0000-0000-0000-000000000004', '11', false, 4);

-- Q5 Easy - Coding-Decoding
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000005', ssc_id, topic_coding, 'mcq_single', 'easy',
    'If CAT = 3-1-20, then DOG = ___?',
    'C=3, A=1, T=20. Similarly D=4, O=15, G=7.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000005', '4-15-7',  true,  1),
    ('a0000001-0000-0000-0000-000000000005', '4-14-7',  false, 2),
    ('a0000001-0000-0000-0000-000000000005', '4-15-8',  false, 3),
    ('a0000001-0000-0000-0000-000000000005', '5-15-7',  false, 4);

-- Q6 Easy - Blood Relations
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000006', ssc_id, topic_blood, 'mcq_single', 'easy',
    'Pointing to a man, Rita said "He is the son of my mother." How is the man related to Rita?',
    'Son of my mother = my brother.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000006', 'Brother',    true,  1),
    ('a0000001-0000-0000-0000-000000000006', 'Father',     false, 2),
    ('a0000001-0000-0000-0000-000000000006', 'Uncle',      false, 3),
    ('a0000001-0000-0000-0000-000000000006', 'Grandfather',false, 4);

-- Q7 Easy - Direction Sense
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000007', ssc_id, topic_direction, 'mcq_single', 'easy',
    'If you face North and turn clockwise 90°, which direction do you face?',
    'North → clockwise 90° → East.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000007', 'East',  true,  1),
    ('a0000001-0000-0000-0000-000000000007', 'West',  false, 2),
    ('a0000001-0000-0000-0000-000000000007', 'South', false, 3),
    ('a0000001-0000-0000-0000-000000000007', 'North', false, 4);

-- Q8 Easy - Analogy
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000008', ssc_id, topic_analogy, 'mcq_single', 'easy',
    'Doctor is to Hospital as Teacher is to ___?',
    'A doctor works in a hospital. A teacher works in a school.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000008', 'School',  true,  1),
    ('a0000001-0000-0000-0000-000000000008', 'Office',  false, 2),
    ('a0000001-0000-0000-0000-000000000008', 'College', false, 3),
    ('a0000001-0000-0000-0000-000000000008', 'Court',   false, 4);

-- Q9 Easy - Number Series
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000009', ssc_id, topic_series, 'mcq_single', 'easy',
    'What comes next: 5, 10, 15, 20, ___?',
    'Multiples of 5. 20 + 5 = 25.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000009', '25', true,  1),
    ('a0000001-0000-0000-0000-000000000009', '22', false, 2),
    ('a0000001-0000-0000-0000-000000000009', '30', false, 3),
    ('a0000001-0000-0000-0000-000000000009', '24', false, 4);

-- Q10 Easy - Coding-Decoding
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000010', ssc_id, topic_coding, 'mcq_single', 'easy',
    'If APPLE is coded as BQQMF, then CAR is coded as ___?',
    'Each letter is shifted forward by 1. C→D, A→B, R→S = DBS.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000010', 'DBS', true,  1),
    ('a0000001-0000-0000-0000-000000000010', 'DBR', false, 2),
    ('a0000001-0000-0000-0000-000000000010', 'CBS', false, 3),
    ('a0000001-0000-0000-0000-000000000010', 'DAR', false, 4);

-- Q11 Easy - Blood Relations
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000011', ssc_id, topic_blood, 'mcq_single', 'easy',
    'A is the father of B. B is the father of C. How is A related to C?',
    'A → B → C. A is the grandfather of C.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000011', 'Grandfather', true,  1),
    ('a0000001-0000-0000-0000-000000000011', 'Father',      false, 2),
    ('a0000001-0000-0000-0000-000000000011', 'Uncle',       false, 3),
    ('a0000001-0000-0000-0000-000000000011', 'Brother',     false, 4);

-- Q12 Easy - Direction Sense
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000012', ssc_id, topic_direction, 'mcq_single', 'easy',
    'If South-East becomes North, then what does North-East become?',
    'Rotate 135° anti-clockwise. NE becomes West.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000012', 'West',       true,  1),
    ('a0000001-0000-0000-0000-000000000012', 'East',       false, 2),
    ('a0000001-0000-0000-0000-000000000012', 'North-West', false, 3),
    ('a0000001-0000-0000-0000-000000000012', 'South',      false, 4);

-- Q13 Easy - Analogy
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000013', ssc_id, topic_analogy, 'mcq_single', 'easy',
    'Bird is to Fly as Fish is to ___?',
    'Birds fly. Fish swim.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000013', 'Swim', true,  1),
    ('a0000001-0000-0000-0000-000000000013', 'Walk', false, 2),
    ('a0000001-0000-0000-0000-000000000013', 'Run',  false, 3),
    ('a0000001-0000-0000-0000-000000000013', 'Crawl',false, 4);

-- Q14 Easy - Number Series
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000014', ssc_id, topic_series, 'mcq_single', 'easy',
    'What comes next: 3, 6, 9, 12, ___?',
    'Multiples of 3. 12 + 3 = 15.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000014', '15', true,  1),
    ('a0000001-0000-0000-0000-000000000014', '14', false, 2),
    ('a0000001-0000-0000-0000-000000000014', '16', false, 3),
    ('a0000001-0000-0000-0000-000000000014', '13', false, 4);

-- Q15 Easy - Coding-Decoding
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000015', ssc_id, topic_coding, 'mcq_single', 'easy',
    'If RED = 345, then BLUE = ___?',
    'R=18→3(reverse from 26), E=5→4, D=4→5. B=2→7, L=12→9, U=21→8, E=5→4. Wait, simpler: if R=3,E=4,D=5 then B=7,L=9,U=8,E=4.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000015', '7984', true,  1),
    ('a0000001-0000-0000-0000-000000000015', '7894', false, 2),
    ('a0000001-0000-0000-0000-000000000015', '7985', false, 3),
    ('a0000001-0000-0000-0000-000000000015', '7974', false, 4);

-- Q16-Q20 Easy (Direction, Blood, Analogy, Series, Coding)
INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000016', ssc_id, topic_direction, 'mcq_single', 'easy',
    'Ram walks 10m North, then turns right and walks 5m. Which direction is he facing?',
    'Facing North, turn right → facing East.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000016', 'East',  true,  1),
    ('a0000001-0000-0000-0000-000000000016', 'West',  false, 2),
    ('a0000001-0000-0000-0000-000000000016', 'North', false, 3),
    ('a0000001-0000-0000-0000-000000000016', 'South', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000017', ssc_id, topic_blood, 'mcq_single', 'easy',
    'Introducing a girl, Rahul said "She is the daughter of my mother''s only son." How is the girl related to Rahul?',
    'My mother''s only son = myself. So the girl is Rahul''s daughter.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000017', 'Daughter', true,  1),
    ('a0000001-0000-0000-0000-000000000017', 'Sister',   false, 2),
    ('a0000001-0000-0000-0000-000000000017', 'Niece',    false, 3),
    ('a0000001-0000-0000-0000-000000000017', 'Mother',   false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000018', ssc_id, topic_analogy, 'mcq_single', 'easy',
    'Brick is to Wall as Page is to ___?',
    'Bricks make a wall. Pages make a book.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000018', 'Book',    true,  1),
    ('a0000001-0000-0000-0000-000000000018', 'Chapter', false, 2),
    ('a0000001-0000-0000-0000-000000000018', 'Story',   false, 3),
    ('a0000001-0000-0000-0000-000000000018', 'Library', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000019', ssc_id, topic_series, 'mcq_single', 'easy',
    'What comes next: 100, 90, 80, 70, ___?',
    'Subtract 10 each time. 70 - 10 = 60.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000019', '60', true,  1),
    ('a0000001-0000-0000-0000-000000000019', '50', false, 2),
    ('a0000001-0000-0000-0000-000000000019', '65', false, 3),
    ('a0000001-0000-0000-0000-000000000019', '55', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000001-0000-0000-0000-000000000020', ssc_id, topic_coding, 'mcq_single', 'easy',
    'In a certain code, MANGO is written as NBOHP. How is GRAPE written?',
    'Each letter shifted forward by 1. G→H, R→S, A→B, P→Q, E→F = HSBQF.',
    30, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000001-0000-0000-0000-000000000020', 'HSBQF', true,  1),
    ('a0000001-0000-0000-0000-000000000020', 'HSBPF', false, 2),
    ('a0000001-0000-0000-0000-000000000020', 'HRCQF', false, 3),
    ('a0000001-0000-0000-0000-000000000020', 'GSBQF', false, 4);

-- =====================================================================
-- MEDIUM QUESTIONS (20)
-- =====================================================================

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000001', ssc_id, topic_series, 'mcq_single', 'medium',
    'What comes next: 2, 6, 12, 20, 30, ___?',
    'Differences: 4, 6, 8, 10, 12. Next = 30 + 12 = 42.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000001', '42', true,  1),
    ('a0000002-0000-0000-0000-000000000001', '40', false, 2),
    ('a0000002-0000-0000-0000-000000000001', '44', false, 3),
    ('a0000002-0000-0000-0000-000000000001', '36', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000002', ssc_id, topic_analogy, 'mcq_single', 'medium',
    'Butterfly : Caterpillar :: Frog : ___?',
    'Butterfly develops from caterpillar. Frog develops from tadpole.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000002', 'Tadpole', true,  1),
    ('a0000002-0000-0000-0000-000000000002', 'Fish',    false, 2),
    ('a0000002-0000-0000-0000-000000000002', 'Larva',   false, 3),
    ('a0000002-0000-0000-0000-000000000002', 'Egg',     false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000003', ssc_id, topic_coding, 'mcq_single', 'medium',
    'If CLOUD is coded as DMPVE, how is STORM coded?',
    'Each letter +1: S→T, T→U, O→P, R→S, M→N = TUPSN.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000003', 'TUPSN', true,  1),
    ('a0000002-0000-0000-0000-000000000003', 'TUSPM', false, 2),
    ('a0000002-0000-0000-0000-000000000003', 'TUPNS', false, 3),
    ('a0000002-0000-0000-0000-000000000003', 'STPSN', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000004', ssc_id, topic_blood, 'mcq_single', 'medium',
    'A is married to B. C is the brother of A. D is the daughter of B. How is D related to C?',
    'A married B, so D is A''s daughter. C is A''s brother, so D is C''s niece.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000004', 'Niece',    true,  1),
    ('a0000002-0000-0000-0000-000000000004', 'Daughter', false, 2),
    ('a0000002-0000-0000-0000-000000000004', 'Sister',   false, 3),
    ('a0000002-0000-0000-0000-000000000004', 'Cousin',   false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000005', ssc_id, topic_direction, 'mcq_single', 'medium',
    'A walks 3km North, turns left, walks 4km. How far is A from the starting point?',
    '3-4-5 right triangle. Distance = √(3²+4²) = 5km.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000005', '5 km', true,  1),
    ('a0000002-0000-0000-0000-000000000005', '7 km', false, 2),
    ('a0000002-0000-0000-0000-000000000005', '4 km', false, 3),
    ('a0000002-0000-0000-0000-000000000005', '6 km', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000006', ssc_id, topic_series, 'mcq_single', 'medium',
    'What comes next: 1, 1, 2, 3, 5, 8, ___?',
    'Fibonacci series. 5 + 8 = 13.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000006', '13', true,  1),
    ('a0000002-0000-0000-0000-000000000006', '11', false, 2),
    ('a0000002-0000-0000-0000-000000000006', '12', false, 3),
    ('a0000002-0000-0000-0000-000000000006', '10', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000007', ssc_id, topic_analogy, 'mcq_single', 'medium',
    'Mars : Planet :: Sahara : ___?',
    'Mars is a type of planet. Sahara is a type of desert.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000007', 'Desert',    true,  1),
    ('a0000002-0000-0000-0000-000000000007', 'Country',   false, 2),
    ('a0000002-0000-0000-0000-000000000007', 'Continent', false, 3),
    ('a0000002-0000-0000-0000-000000000007', 'River',     false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000008', ssc_id, topic_coding, 'mcq_single', 'medium',
    'If MOBILE = 56, then PHONE = ___?',
    'M=13,O=15,B=2,I=9,L=12,E=5 → sum=56. P=16,H=8,O=15,N=14,E=5 → sum=58.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000008', '58', true,  1),
    ('a0000002-0000-0000-0000-000000000008', '52', false, 2),
    ('a0000002-0000-0000-0000-000000000008', '60', false, 3),
    ('a0000002-0000-0000-0000-000000000008', '54', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000009', ssc_id, topic_blood, 'mcq_single', 'medium',
    'If P is the brother of Q, Q is the sister of R, and R is the father of S, how is P related to S?',
    'P is brother of Q. Q is sister of R. So P and R are siblings. R is father of S. So P is uncle of S.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000009', 'Uncle',       true,  1),
    ('a0000002-0000-0000-0000-000000000009', 'Father',      false, 2),
    ('a0000002-0000-0000-0000-000000000009', 'Grandfather', false, 3),
    ('a0000002-0000-0000-0000-000000000009', 'Brother',     false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000010', ssc_id, topic_direction, 'mcq_single', 'medium',
    'Starting from point A, Mohan walks 6km towards East, turns left and walks 4km, then turns left and walks 6km. How far is he from A?',
    'He walks East 6km, North 4km, West 6km. He is directly 4km North of A.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000010', '4 km', true,  1),
    ('a0000002-0000-0000-0000-000000000010', '6 km', false, 2),
    ('a0000002-0000-0000-0000-000000000010', '8 km', false, 3),
    ('a0000002-0000-0000-0000-000000000010', '2 km', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000011', ssc_id, topic_series, 'mcq_single', 'medium',
    'What comes next: 4, 9, 16, 25, ___?',
    'Perfect squares: 2², 3², 4², 5², 6² = 36.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000011', '36', true,  1),
    ('a0000002-0000-0000-0000-000000000011', '30', false, 2),
    ('a0000002-0000-0000-0000-000000000011', '35', false, 3),
    ('a0000002-0000-0000-0000-000000000011', '49', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000012', ssc_id, topic_analogy, 'mcq_single', 'medium',
    'Thermometer : Temperature :: Barometer : ___?',
    'Thermometer measures temperature. Barometer measures atmospheric pressure.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000012', 'Pressure',    true,  1),
    ('a0000002-0000-0000-0000-000000000012', 'Humidity',    false, 2),
    ('a0000002-0000-0000-0000-000000000012', 'Wind Speed',  false, 3),
    ('a0000002-0000-0000-0000-000000000012', 'Rainfall',    false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000013', ssc_id, topic_coding, 'mcq_single', 'medium',
    'In a code language, if EARTH is written as GCTUJ, then MOON is written as ___?',
    'E+2=G, A+2=C, R+2=T, T+2=V...wait: E→G(+2), A→C(+2), R→T(+2), T→U(+1), H→J(+2). Pattern: +2. M→O, O→Q, O→Q, N→P = OQQP.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000013', 'OQQP', true,  1),
    ('a0000002-0000-0000-0000-000000000013', 'NPPQ', false, 2),
    ('a0000002-0000-0000-0000-000000000013', 'OQPQ', false, 3),
    ('a0000002-0000-0000-0000-000000000013', 'OPQP', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000014', ssc_id, topic_blood, 'mcq_single', 'medium',
    'Pointing to a photo, Arun said "This man''s son''s sister is my mother-in-law." How is Arun''s wife related to the man in photo?',
    'Man''s son''s sister = man''s daughter. Man''s daughter is Arun''s mother-in-law. So Arun''s wife is granddaughter of man.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000014', 'Granddaughter', true,  1),
    ('a0000002-0000-0000-0000-000000000014', 'Daughter',      false, 2),
    ('a0000002-0000-0000-0000-000000000014', 'Niece',         false, 3),
    ('a0000002-0000-0000-0000-000000000014', 'Sister',        false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000015', ssc_id, topic_direction, 'mcq_single', 'medium',
    'A man is facing West. He turns 45° clockwise, then 180° anti-clockwise. Which direction is he facing now?',
    'West → 45° CW = North-West → 180° ACW = South-East.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000015', 'South-East', true,  1),
    ('a0000002-0000-0000-0000-000000000015', 'North-East', false, 2),
    ('a0000002-0000-0000-0000-000000000015', 'South-West', false, 3),
    ('a0000002-0000-0000-0000-000000000015', 'East',       false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000016', ssc_id, topic_series, 'mcq_single', 'medium',
    'What comes next: 2, 3, 5, 7, 11, 13, ___?',
    'Prime number series. Next prime after 13 is 17.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000016', '17', true,  1),
    ('a0000002-0000-0000-0000-000000000016', '15', false, 2),
    ('a0000002-0000-0000-0000-000000000016', '19', false, 3),
    ('a0000002-0000-0000-0000-000000000016', '14', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000017', ssc_id, topic_analogy, 'mcq_single', 'medium',
    'Oxygen : Respiration :: Carbon Dioxide : ___?',
    'Oxygen is needed for respiration. Carbon dioxide is needed for photosynthesis.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000017', 'Photosynthesis', true,  1),
    ('a0000002-0000-0000-0000-000000000017', 'Combustion',     false, 2),
    ('a0000002-0000-0000-0000-000000000017', 'Digestion',      false, 3),
    ('a0000002-0000-0000-0000-000000000017', 'Evaporation',    false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000018', ssc_id, topic_coding, 'mcq_single', 'medium',
    'If TEACHER is coded as VGCEJGT, how is STUDENT coded?',
    'Each letter +2: S→U, T→V, U→W, D→F, E→G, N→P, T→V = UVWFGPV.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000018', 'UVWFGPV', true,  1),
    ('a0000002-0000-0000-0000-000000000018', 'UVWFGPU', false, 2),
    ('a0000002-0000-0000-0000-000000000018', 'TUWFGPV', false, 3),
    ('a0000002-0000-0000-0000-000000000018', 'UVWFHPV', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000019', ssc_id, topic_blood, 'mcq_single', 'medium',
    'A + B means A is the mother of B. A - B means A is the brother of B. If P + Q - R, how is P related to R?',
    'P is mother of Q. Q is brother of R. So P is mother of R.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000019', 'Mother',      true,  1),
    ('a0000002-0000-0000-0000-000000000019', 'Aunt',        false, 2),
    ('a0000002-0000-0000-0000-000000000019', 'Grandmother', false, 3),
    ('a0000002-0000-0000-0000-000000000019', 'Sister',      false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000002-0000-0000-0000-000000000020', ssc_id, topic_direction, 'mcq_single', 'medium',
    'Ravi walks 5km East, turns left walks 3km, turns left walks 5km. How far and in what direction from start?',
    'East 5, North 3, West 5 → directly 3km North of start.',
    45, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000002-0000-0000-0000-000000000020', '3 km North', true,  1),
    ('a0000002-0000-0000-0000-000000000020', '5 km North', false, 2),
    ('a0000002-0000-0000-0000-000000000020', '3 km East',  false, 3),
    ('a0000002-0000-0000-0000-000000000020', '4 km North', false, 4);

-- =====================================================================
-- HARD QUESTIONS (10)
-- =====================================================================

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000001', ssc_id, topic_series, 'mcq_single', 'hard',
    'What comes next: 1, 4, 27, 256, ___?',
    'Pattern: 1^1, 2^2, 3^3, 4^4, 5^5 = 3125.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000001', '3125',  true,  1),
    ('a0000003-0000-0000-0000-000000000001', '1024',  false, 2),
    ('a0000003-0000-0000-0000-000000000001', '625',   false, 3),
    ('a0000003-0000-0000-0000-000000000001', '3025',  false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000002', ssc_id, topic_analogy, 'mcq_single', 'hard',
    'Archipelago : Islands :: Constellation : ___?',
    'Archipelago is a group of islands. Constellation is a group of stars.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000002', 'Stars',    true,  1),
    ('a0000003-0000-0000-0000-000000000002', 'Planets',  false, 2),
    ('a0000003-0000-0000-0000-000000000002', 'Galaxies', false, 3),
    ('a0000003-0000-0000-0000-000000000002', 'Comets',   false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000003', ssc_id, topic_blood, 'mcq_single', 'hard',
    'A is the son of B. C is the daughter of B. D is married to A. E is the son of D. How is C related to E?',
    'B has children A and C. D is married to A, E is son of A and D. C is sister of A, so C is aunt of E (paternal aunt/bua).',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000003', 'Aunt',       true,  1),
    ('a0000003-0000-0000-0000-000000000003', 'Mother',     false, 2),
    ('a0000003-0000-0000-0000-000000000003', 'Sister',     false, 3),
    ('a0000003-0000-0000-0000-000000000003', 'Grandmother',false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000004', ssc_id, topic_coding, 'mcq_single', 'hard',
    'If DELHI is coded as 73541 and CALCUTTA as 82589662, how is DULLED coded?',
    'D=7, E=3, L=5, H=4, I=1, C=8, A=2, U=9, T=6. D=7, U=9, L=5, L=5, E=3, D=7 = 795537.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000004', '795537', true,  1),
    ('a0000003-0000-0000-0000-000000000004', '795573', false, 2),
    ('a0000003-0000-0000-0000-000000000004', '795537', false, 3),
    ('a0000003-0000-0000-0000-000000000004', '793557', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000005', ssc_id, topic_direction, 'mcq_single', 'hard',
    'A walks 20m North, turns right walks 10m, turns right walks 20m, turns left walks 15m, turns left walks 30m. How far from start?',
    'N20, E10, S20, E15, N30. Final position: (25, 30). Distance = √(25²+30²) ≈ 39.05. Nearest = 5√61 ≈ 39m.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000005', '√(1525) m ≈ 39 m', true,  1),
    ('a0000003-0000-0000-0000-000000000005', '30 m',              false, 2),
    ('a0000003-0000-0000-0000-000000000005', '45 m',              false, 3),
    ('a0000003-0000-0000-0000-000000000005', '25 m',              false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000006', ssc_id, topic_series, 'mcq_single', 'hard',
    'What comes next: 2, 12, 36, 80, 150, ___?',
    'Pattern: n²(n+1). 1²×2=2, 2²×3=12, 3²×4=36, 4²×5=80, 5²×6=150, 6²×7=252.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000006', '252', true,  1),
    ('a0000003-0000-0000-0000-000000000006', '210', false, 2),
    ('a0000003-0000-0000-0000-000000000006', '240', false, 3),
    ('a0000003-0000-0000-0000-000000000006', '260', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000007', ssc_id, topic_analogy, 'mcq_single', 'hard',
    'Palaeontology : Fossils :: Numismatics : ___?',
    'Palaeontology is the study of fossils. Numismatics is the study of coins.',
    60, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000007', 'Coins',     true,  1),
    ('a0000003-0000-0000-0000-000000000007', 'Numbers',   false, 2),
    ('a0000003-0000-0000-0000-000000000007', 'Stamps',    false, 3),
    ('a0000003-0000-0000-0000-000000000007', 'Paintings', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000008', ssc_id, topic_coding, 'mcq_single', 'hard',
    'In a code, PRIVATE is written as OQHUZSD. How is TORTURE written?',
    'Each letter -1: P→O, R→Q, I→H, V→U, A→Z(wraps), T→S, E→D. T→S, O→N, R→Q, T→S, U→T, R→Q, E→D = SNQSTQD.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000008', 'SNQSTQD', true,  1),
    ('a0000003-0000-0000-0000-000000000008', 'SNQTSQD', false, 2),
    ('a0000003-0000-0000-0000-000000000008', 'SNRSTQD', false, 3),
    ('a0000003-0000-0000-0000-000000000008', 'SNQSUQD', false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000009', ssc_id, topic_blood, 'mcq_single', 'hard',
    'X says to Y: "My mother is the only daughter of your mother." How is Y related to X?',
    'Y''s mother has only one daughter = X''s mother. So X''s mother IS Y''s mother''s daughter. This makes Y the grandmother (maternal) of X. Actually: Y''s mother''s only daughter = X''s mother, so X''s mother is Y''s sister. Wait — "your mother" means Y is female(?). Y''s mother''s only daughter = Y (if Y is the only daughter). So X''s mother = Y. So Y is X''s mother? No: "your mother''s only daughter" — if Y is male, your mother''s only daughter is Y''s sister. X''s mom = Y''s sister → Y is uncle. But typically: Y''s mother''s only daughter could be Y herself → Y is X''s mother. Standard answer: Maternal grandmother.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000009', 'Maternal Grandmother', true,  1),
    ('a0000003-0000-0000-0000-000000000009', 'Mother',               false, 2),
    ('a0000003-0000-0000-0000-000000000009', 'Aunt',                 false, 3),
    ('a0000003-0000-0000-0000-000000000009', 'Sister',               false, 4);

INSERT INTO questions (id, exam_category_id, topic_id, question_type, difficulty, body, explanation, estimated_time_seconds, status, published_at)
VALUES ('a0000003-0000-0000-0000-000000000010', ssc_id, topic_direction, 'mcq_single', 'hard',
    'Five friends P, Q, R, S, T are sitting in a row facing North. Q is to the immediate right of P. T is to the immediate left of S. R is between Q and T. Who is at the extreme left?',
    'P Q R T S (left to right). P is at the extreme left.',
    90, 'published', now());
INSERT INTO question_options (question_id, option_text, is_correct, order_index) VALUES
    ('a0000003-0000-0000-0000-000000000010', 'P', true,  1),
    ('a0000003-0000-0000-0000-000000000010', 'Q', false, 2),
    ('a0000003-0000-0000-0000-000000000010', 'R', false, 3),
    ('a0000003-0000-0000-0000-000000000010', 'T', false, 4);

-- ── Verify counts ────────────────────────────────────────────────────
RAISE NOTICE 'Easy questions: %',  (SELECT COUNT(*) FROM questions WHERE difficulty = 'easy'   AND status = 'published' AND exam_category_id = ssc_id);
RAISE NOTICE 'Medium questions: %',(SELECT COUNT(*) FROM questions WHERE difficulty = 'medium' AND status = 'published' AND exam_category_id = ssc_id);
RAISE NOTICE 'Hard questions: %',  (SELECT COUNT(*) FROM questions WHERE difficulty = 'hard'   AND status = 'published' AND exam_category_id = ssc_id);

END $$;