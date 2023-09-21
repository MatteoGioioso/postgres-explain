const fs = require('fs').promises

async function assignColor() {
  const file = await fs.readFile('../wait_events.json')
  const we = JSON.parse(file)
  for (const weKey in we) {
    we[weKey].color = pastelColors()
  }

  await fs.writeFile('../wait_events.json',JSON.stringify(we))
}

function pastelColors() {
  const r = (Math.round(Math.random() * 127) + 127).toString(16);
  const g = (Math.round(Math.random() * 127) + 127).toString(16);
  const b = (Math.round(Math.random() * 127) + 127).toString(16);
  return '#' + r + g + b;
}

assignColor().then()