require "fileutils"
require "tempfile"

module Kenban
  class Store
    attr_reader :path

    def initialize(path = nil)
      @path = path || self.class.default_path
    end

    def self.default_path
      local = File.join(Dir.pwd, "tasks.txt")
      return local if File.exist?(local)

      home_dir = File.join(Dir.home, ".kenban")
      File.join(home_dir, "tasks.txt")
    end

    # Read all tasks from the file. Returns an array of Task objects.
    def read_tasks
      ensure_file_exists
      lines = File.readlines(path, chomp: true)
      tasks = []
      lines.each_with_index do |line, i|
        next if line.strip.empty?
        begin
          task = Task.parse(line)
          tasks << task if task
        rescue ArgumentError => e
          warn "Warning: line #{i + 1}: #{e.message}"
        end
      end
      tasks
    end

    # Write tasks back to the file atomically.
    def write_tasks(tasks)
      ensure_dir_exists
      dir = File.dirname(path)
      tmp = Tempfile.new("kenban", dir)
      begin
        tasks.each { |t| tmp.puts(t.to_s) }
        tmp.close
        File.rename(tmp.path, path)
      rescue
        tmp.close
        tmp.unlink
        raise
      end
    end

    # Append a single task to the file.
    def append_task(task)
      ensure_file_exists
      File.open(path, "a") { |f| f.puts(task.to_s) }
    end

    # Replace a task at a given 0-based index.
    def replace_task(index, new_task)
      tasks = read_tasks
      unless index >= 0 && index < tasks.size
        raise ArgumentError, "Task number out of range (1..#{tasks.size})"
      end
      tasks[index] = new_task
      write_tasks(tasks)
    end

    private

    def ensure_file_exists
      ensure_dir_exists
      FileUtils.touch(path) unless File.exist?(path)
    end

    def ensure_dir_exists
      dir = File.dirname(path)
      FileUtils.mkdir_p(dir) unless Dir.exist?(dir)
    end
  end
end
