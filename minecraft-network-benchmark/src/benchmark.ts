import mineflayer from "mineflayer";
import { Vec3 } from "vec3";

const VERSION = "1.21.4";
const host = "localhost"; // Replace with your proxy server's host
const port = 25565; // Replace with your proxy server's port
const botCount = 5; // Number of bots to create

async function createBot(username: string, host: string, port: number) {
  return new Promise<void>((resolve, reject) => {
    const bot = mineflayer.createBot({
      host: host,
      port: port,
      username: username,
      version: VERSION,
      logErrors: true,
      hideErrors: false,
    });

    bot.on("spawn", () => {
      // console.log(`${username} spawned`);
      // Uncomment the following lines to simulate movement and chat
      simulateMovement(bot);
      simulateChat(bot);
    });

    bot.once("login", () => {
      console.log(`${username} logged in`);
      // simulateMovement(bot);
      // simulateChat(bot);
      resolve();
    });

    bot.on("chat", (username, message) => {
      // console.log(`${username}: ${message}`);
    });

    bot.on("kicked", (reason) => {
      console.log(`${username} was kicked: ${reason}`);
      reject(new Error(`Bot ${username} was kicked: ${reason}`));
    });

    bot.on("end", (reason) => {
      console.log(`${username} was kicked: ${reason}`);
      reject(new Error(`Bot ${username} was kicked: ${reason}`));
    });

    bot.on("error", (err) => {
      console.log(`${username} encountered an error: ${err}`);
      reject(err);
    });
  });
}

function simulateMovement(bot: mineflayer.Bot) {
  setInterval(() => {
    const x = bot.entity.position.x + (Math.random() * 2 - 1);
    const y = bot.entity.position.y;
    const z = bot.entity.position.z + (Math.random() * 2 - 1);
    bot.lookAt(new Vec3(x, y, z), true);
    bot.setControlState("forward", true);
    setTimeout(() => {
      bot.setControlState("forward", false);
    }, 1000);
  }, 2000);
}

function simulateChat(bot: mineflayer.Bot) {
  setInterval(() => {
    const messages = [
      "Hello!",
      "How are you?",
      "This is a test message.",
      "Benchmarking the server.",
      "PrismarineJS is awesome!",
    ];
    const message = messages[Math.floor(Math.random() * messages.length)];
    bot.chat(message);
  }, 5000);
}

async function createBots(count: number, host: string, port: number) {
  for (let i = 0; i < count; i++) {
    const username = `Bot${i + 1}`;
    try {
      await createBot(username, host, port);
      console.log(`Bot ${username} created successfully`);
    } catch (err) {
      console.error(`Failed to create bot ${username}:`, err);
    }
    // Add a delay between bot creations to avoid overwhelming the server
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
}

createBots(botCount, host, port)
  .then(() => {
    console.log("All bots created");
  })
  .catch((err) => {
    console.error("Error creating bots:", err);
  });
