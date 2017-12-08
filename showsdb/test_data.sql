--
-- Name: episodes; Type: TABLE; Schema: public; Owner: shows
--

CREATE TABLE episodes (
    id integer NOT NULL,
    show integer,
    season integer,
    episode integer
);

CREATE TABLE options (
    id integer NOT NULL,
    name character varying(255),
    value character varying(255)
);

--
-- Name: shows; Type: TABLE; Schema: public; Owner: shows
--

CREATE TABLE shows (
    id integer NOT NULL,
    name character varying(255),
    active boolean DEFAULT true
);

COPY episodes (id, show, season, episode) FROM stdin;
1	1	1	1
2	1	1	2
3	1	1	3
4	1	2	1
5	1	2	2
6	1	2	3
7	2	1	1
8	2	1	2
9	2	1	3
10	2	2	1
11	2	2	2
12	2	2	3
13	3	1	1
14	3	1	2
15	3	1	3
16	3	2	1
17	3	2	2
18	3	2	3
\.

--
-- Data for Name: shows; Type: TABLE DATA; Schema: public; Owner: shows
--

COPY shows (id, name, active) FROM stdin;
1	Test Show 1	t
2	Test Show 2	t
3	Test Show 3	t
\.

