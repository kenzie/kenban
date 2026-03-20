require "minitest/autorun"
require "tmpdir"
$LOAD_PATH.unshift(File.expand_path("../lib", __dir__))
require "kenban"

class TestStore < Minitest::Test
  def setup
    @dir = Dir.mktmpdir("kenban-test")
    @path = File.join(@dir, "tasks.txt")
    @store = Kenban::Store.new(@path)
  end

  def teardown
    FileUtils.rm_rf(@dir)
  end

  def test_read_empty_file
    tasks = @store.read_tasks
    assert_equal [], tasks
  end

  def test_append_and_read
    task = Kenban::Task.new(state: "todo", project: "proj", description: "Do thing")
    @store.append_task(task)

    tasks = @store.read_tasks
    assert_equal 1, tasks.size
    assert_equal "todo", tasks[0].state
    assert_equal "proj", tasks[0].project
    assert_equal "Do thing", tasks[0].description
  end

  def test_write_tasks_atomically
    tasks = [
      Kenban::Task.new(state: "todo", project: "a", description: "First"),
      Kenban::Task.new(state: "doing", project: "b", description: "Second"),
    ]
    @store.write_tasks(tasks)

    read_back = @store.read_tasks
    assert_equal 2, read_back.size
    assert_equal "[todo] [a] First", read_back[0].to_s
    assert_equal "[doing] [b] Second", read_back[1].to_s
  end

  def test_creates_file_if_missing
    new_path = File.join(@dir, "sub", "tasks.txt")
    store = Kenban::Store.new(new_path)
    tasks = store.read_tasks
    assert_equal [], tasks
    assert File.exist?(new_path)
  end

  def test_blank_lines_ignored
    File.write(@path, "[todo] [p] One\n\n  \n[done] [p] Two\n")
    tasks = @store.read_tasks
    assert_equal 2, tasks.size
  end

  def test_malformed_lines_warned_not_crashed
    File.write(@path, "[todo] [p] Good\nbad line\n[done] [p] Also good\n")
    out, _err = capture_io { @store.read_tasks }
    # should still get 2 valid tasks
  end

  def test_replace_task
    File.write(@path, "[todo] [a] First\n[todo] [b] Second\n[todo] [c] Third\n")
    new_task = Kenban::Task.new(state: "doing", project: "b", description: "Second updated")
    @store.replace_task(1, new_task)

    tasks = @store.read_tasks
    assert_equal "doing", tasks[1].state
    assert_equal "Second updated", tasks[1].description
    # others unchanged
    assert_equal "First", tasks[0].description
    assert_equal "Third", tasks[2].description
  end
end
