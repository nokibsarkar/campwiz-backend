module.exports = {
    apps: [{
        name: "campwiz-backend-1",
        script: "./campwiz",
        args: "-port 8081"
    }, {
        name: "campwiz-backend-2",
        script: "./campwiz",
        args: "-port 8082"
    }]
}
