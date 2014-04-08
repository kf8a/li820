require 'serialport'
require 'nokogiri'
require 'ffi-rzmq'
require 'json'

class LI820

  MODEL = "li820"

  def sample()
    buffer = ""
    open do |sp|
			buffer << waiting(sp)
			buffer << reading(sp)
    end
    parse('<root>'+buffer+'</root>')
  end

  protected

  def waiting(sp)
		buffer = ''
		while not buffer =~ /<li820>/
			buffer << sp.read(1)
		end
		"<li840>"	
  end

	def reading(sp)
		buffer = ''
    while not buffer =~ /<\/li820>/
			buffer << sp.read(1)
		end
		buffer
	end

  def parse(buffer="")
    doc = Nokogiri::XML(buffer)
    co2 = doc.search("//data/co2").text.to_f
    h2o = doc.search("//data/h2o").text.to_f

    {at: Time.now, co2: co2, h2o: h2o}
  end

  def open
    port_str = "/dev/ttyS4"
    baud_rate = 9600
    data_bits = 8
    stop_bits = 1
    parity = SerialPort::NONE

    sp = SerialPort.new(port_str, baud_rate, data_bits, stop_bits, parity)

    yield(sp)

    sp.close
  end
end

class LIServer
  def run(li820)
    context = ZMQ::Context.new
    publisher = context.socket(ZMQ::PUB)
    publisher.bind("tcp://*:5556")
    publisher.bind("ipc://weather.ipc")

    while true
      update = JSON.generate(li820.sample)
      # puts update 
      publisher.send_string(update)
    end
  end
end

if __FILE__ == $0
  li820 = LI820.new
  server = LIServer.new
  server.run(li820)
end
