var canvas, ctx;
var colour = "black";
var lineThickness = 20;
var isPainting = false;
var startX = 0;
var startY = 0;
            
async function init() {
    console.log("init");
    canvas = document.getElementById('mask');
    ctx = canvas.getContext("2d");
    w = canvas.width;
    h = canvas.height;
    try {
        await fetchMask();
    } catch (error) {
        console.error(error);
        reset();
    }

    canvas.addEventListener("mousemove", (e) => draw(e));
    canvas.addEventListener("mousedown", function (e) {
        console.log("mousedown");
        isPainting = true;
        // startX = e.clientX - canvas.offsetLeft;
        // startY = e.clientY - canvas.offsetTop;
    });
    canvas.addEventListener("mouseup", function (e) {
        console.log("mouseup");
        isPainting = false;
        ctx.stroke();
        ctx.beginPath();
        sendMaskUpdate();
    }, false);
    canvas.addEventListener("mouseout", function (e) {
        isPainting = false;
    });
}

function draw(e) {
    if (!isPainting) return;
    console.log("draw");
    ctx.globalCompositeOperation = 'destination-out';
    ctx.lineWidth = lineThickness;
    ctx.lineCap = 'round';

    console.log("drawing to " + (e.clientX - canvas.offsetLeft) + ", " + (e.clientY - canvas.offsetTop));
    ctx.lineTo(e.clientX - canvas.offsetLeft, e.clientY - canvas.offsetTop);
    ctx.stroke();
}

async function fetchMask() {
    console.log("fetching mask");
    var img = new Image();
    var p = new Promise((resolve, reject) => {
        img.onload = function() {
            console.log("image loaded");
            ctx.drawImage(img, 0, 0);
            resolve();
        };
        img.onerror = function(err) {
            console.error("image failed to load", err);
            reject(err);
        };
        img.src = 'http://localhost:8080/mask';
    });
    return p;
}

async function reset() {
    ctx.globalCompositeOperation = 'source-over';
    ctx.beginPath();
    ctx.strokeStyle = colour;
    ctx.fillRect(0, 0, w, h);
    ctx.stroke();
    await sendMaskUpdate();
}

async function reveal() {
    ctx.globalCompositeOperation = 'destination-out';
    ctx.beginPath();
    ctx.strokeStyle = colour;
    ctx.fillRect(0, 0, w, h);
    ctx.stroke();
    await sendMaskUpdate();
}

async function sendMaskUpdate() {
    canvas.toBlob(async (blob) => {
        console.log("blob created");
        console.log("sending mask update");
        await fetch('/mask', {
            method: 'PUT',
            headers: {
                'Content-Type': 'image/png'
            },
            body: blob
        })
        .then(response => console.log(response))
    });
    
}