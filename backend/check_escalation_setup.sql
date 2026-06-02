-- Check if escalation tables exist and diagnose issues
-- Run this to verify your escalation system setup

-- 1. Check if escalation tables exist
SELECT 
    'complaint_escalations' as table_name,
    CASE WHEN EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'complaint_escalations'
    ) THEN '✅ EXISTS' ELSE '❌ MISSING - Run complaint_escalation_schema.sql' END as status
UNION ALL
SELECT 
    'complaint_escalation_history' as table_name,
    CASE WHEN EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'complaint_escalation_history'
    ) THEN '✅ EXISTS' ELSE '❌ MISSING - Run complaint_escalation_schema.sql' END as status;

-- 2. Check table structures (if they exist)
SELECT 
    column_name, 
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'complaint_escalations'
ORDER BY ordinal_position;

-- 3. Count existing escalations
SELECT 
    COUNT(*) as total_escalations,
    COUNT(CASE WHEN next_escalation_due IS NOT NULL THEN 1 END) as with_due_date,
    COUNT(CASE WHEN next_escalation_due <= NOW() THEN 1 END) as overdue
FROM complaint_escalations
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'complaint_escalations');

-- 4. Check complaints without escalation
SELECT 
    COUNT(*) as complaints_without_escalation
FROM report_issues ri
WHERE NOT EXISTS (
    SELECT 1 FROM complaint_escalations ce 
    WHERE ce.complaint_id = ri.id
)
AND ri.status NOT IN ('resolved', 'closed')
AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'complaint_escalations');

-- 5. Sample escalation data (if exists)
SELECT 
    ce.complaint_id,
    ce.current_level,
    ce.current_team,
    ce.next_escalation_due,
    ri.status,
    ri.issue_type,
    ri.created_at as complaint_created
FROM complaint_escalations ce
JOIN report_issues ri ON ce.complaint_id = ri.id
ORDER BY ce.created_at DESC
LIMIT 5;
