/*
Package problem is a package for test problem manipulation.

note that the following docuemnts are out of date.

# Tests

Usually a problem contains multiple testcases, whose data are stored in
dir/data/tests.  All testcases have the same fields.

# Subtask

To better assign points to testcases with different intensity, it's common to
set up several subtasks for the problem, each containing a series of testcases.

Note that if subtask is enabled, independent testcases (i.e. not in a subtask)
are not allowed.

If subtask data occured (at least one field, at least one record), subtask is
enabled.

For some problem, different subtasks use different files to test contestant's
submission, for example, checker or input generator. Thus these data are stored
in dir/data/tests/. Again, all subtasks' data have the same fields.

Subtask may have "_score" field, whose value is a number string.

For problem enabling subtask, its testcase and subtask both contain
"_subtaskid" field, determining which subtask the testcase belongs to.

# Static Data

Common data are stored in dir/data/static/ shared by all testcases.

# Testcase Score

Score of a testcase is calculated as follows:

If subtask is enabled, testcase's "_score" is ignored, its score is
{subtask score} / {number of tests in this subtask}.

Otherwise, if "average" is specified for "_score" field, its score is
{problem total score} / {number of testcases}. Else its "_score" should be
a number string, denoting its score.

# Statement

Markdown is the standard format of statement. Other formats may be supported in
the future.

# Workflow

See github.com/super-yaoj/yaoj-core/workflow.

Problem gives workflow 3 datagroups naming "testcase", "static" and
"submission" respectively.

# Hack

hack 是指对选手代码的纠错。hack 的通常流程为

1. 指定一份提交为标准答案

2. 提供一份（可能不完整）的数据执行 workflow，跳过无法执行的结点，这样可以得到一些中间输出文件

3. 利用一部分中间输出文件来填补数据后执行选手的代码，判断是否正确。
*/
package problem
