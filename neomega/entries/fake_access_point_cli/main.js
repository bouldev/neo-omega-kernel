const zmq=require("zeromq");

function parseHaxHex(str) {
	return Buffer.from(str.replace(/::(.*?)::/g, (a,b)=>{
		let buf=Buffer.from("00"+b);
		buf.writeUInt16LE(b.length);
		return buf.toString("hex");
	}).replace(/(#(.*?)(\n|$)|( |\t|\n)+)/g, "").replace(
		/\[\[(.*?)\]\]/g,
		(a,b)=>{
			let size_buf=Buffer.alloc(8);
			size_buf.writeBigUInt64LE(BigInt(b.length/2));
			return size_buf.toString("hex")+b;
		}
	),"hex");
}

function processMessage(msg) {
	if(msg=="botReady") {
		return "botReady"
	}else if(msg=="getConnShieldID") {
		return Buffer.alloc(4);
	}else if(msg=="getUQHolderBytes") {
		return parseHaxHex(`
		::BotBasicInfoHolder::
		[[
			::NetSucks:: # BotName
		deadbeeffeedfeed # runtime id
		feedfeeddeadbeef # unique id
			::WhoCares:: # BotIdentity
		]]
		::PlayersInfoHolder::
		[[
		00 00 00 00 # PlayersLen = no player
		]]
		::ExtendInfo::
		[[
		00 00 # CompressThreshold
		00 # KnownCompressThreshold
		01 00 00 00 # WorldGameMode
		01 # KnownWorldGameMode
		00 00 00 00 # WorldDifficulty
		01 # KnownWorldDifficulty
		00 00 00 00 # Time
		01 # KnownTime
		00 00 00 00 # DayTime
		01 # KnownDayTime
		00 00 00 00 # DayTimePercent
		01 # KnownDayTimePercent
		00 00 00 00 # Length of gamerules
		01 # KnownGameRules
		]]`);
	}
}

async function main() {
	const sock=new zmq.Publisher;
	const another_sock=new zmq.Dealer;
	await another_sock.bind("tcp://127.0.0.1:24015");
	await sock.bind("tcp://127.0.0.1:24016");
	for await (const msg of another_sock) {
		console.log("Message: %s %s", msg[0], msg[1]);
		let v=processMessage(msg[1]);
		if(v) {
			another_sock.send([msg[0],v]);
		}
		//another_sock.send(msg);
	}
}
main();
