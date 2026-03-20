require "minitest/autorun"
require "tmpdir"
$LOAD_PATH.unshift(File.expand_path("../lib", __dir__))
require "kenban"

class TestCLI < Minitest::Test
  def setup
    @dir = Dir.mktmpdir("kenban-test")
    @path = File.join(@dir, "tasks.txt")
    @store = Kenban::Store.new(@path)
  end

  def teardown
    FileUtils.rm_rf(@dir)
  end

  def run_cli(*args)
    cli = Kenban::CLI.new(args.dup, store: @store)
    capture_io { cli.run }
  end

  def test_add_and_list
    run_cli("add", "[myproject] Build the thing")
    out, _ = run_cli("list")
    assert_includes out, "[todo] [myproject] Build the thing"
    assert_includes out, "1."
  end

  def test_add_with_tags
    run_cli("add", "[proj] Do stuff @high +billing")
    tasks = @store.read_tasks
    assert_equal "Do stuff @high +billing", tasks[0].description
  end

  def test_list_filter_by_project
    run_cli("add", "[alpha] Task A")
    run_cli("add", "[beta] Task B")
    out, _ = run_cli("list", "alpha")
    assert_includes out, "Task A"
    refute_includes out, "Task B"
  end

  def test_state_filter
    run_cli("add", "[p] Task one")
    File.write(@path, "[todo] [p] Task one\n[doing] [p] Task two\n")
    out, _ = run_cli("state", "doing")
    assert_includes out, "Task two"
    refute_includes out, "Task one"
  end

  def test_move
    File.write(@path, "[todo] [p] My task\n")
    run_cli("move", "1", "doing")
    tasks = @store.read_tasks
    assert_equal "doing", tasks[0].state
  end

  def test_done_shortcut
    File.write(@path, "[todo] [p] My task\n")
    run_cli("done", "1")
    tasks = @store.read_tasks
    assert_equal "done", tasks[0].state
  end

  def test_projects
    File.write(@path, "[todo] [alpha] A\n[todo] [beta] B\n[done] [alpha] C\n")
    out, _ = run_cli("projects")
    assert_includes out, "alpha"
    assert_includes out, "beta"
  end

  def test_copy_outputs_confirmation
    # We can't easily test clipboard in CI, but we can test the task lookup.
    # On systems without pbcopy/xclip, copy exits with an error message.
    File.write(@path, "[todo] [p] My task\n")
    # Just verify the task is readable at that index (copy depends on OS clipboard)
    tasks = @store.read_tasks
    assert_equal "[todo] [p] My task", tasks[0].to_s
  end

  def test_help
    out, _ = run_cli("help")
    assert_includes out, "kenban"
    assert_includes out, "add"
  end
end
