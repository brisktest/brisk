# lets kill before OOM does it for us
# need a memory checker in a thread that sends a SIGTERM to the current process if memory exceeds max

if ENV['MEM_KILLER']
  Thread.abort_on_exception = true
  max_mem = ENV['MEM_KILLER_MAX_MEM'] || 4_294_967_296
  max_mem_bytes = (rand + 9) * max_mem.to_i / 10
  Rails.logger.debug { "MEM_KILLER: max mem for this thread is #{max_mem_bytes}" }
  require 'os'

  t = Thread.new do
    while true
      # adding some randomness here so that all our processes don't die at the same time.
      if OS.rss_bytes > max_mem_bytes
        pid = Process.pid
        Rails.logger.warn("MEM_KILLER: killing process pid: #{pid} it has hit rss #{OS.rss_bytes} which is greater than our limit of #{max_mem_bytes}")
        Sentry.capture_message("MEM_KILLER: killing process pid: #{pid} it has hit rss #{OS.rss_bytes} which is greater than our limit of #{max_mem_bytes}")
        Process.kill('TERM', pid)
      else
        Rails.logger.debug("Memory is #{OS.rss_bytes}")
        sleep ENV['MEM_KILLER_SLEEP_TIME'] || 60
      end
    end
  end
end
