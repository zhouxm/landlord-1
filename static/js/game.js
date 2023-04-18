
PG.Game = function(game) {

    this.roomId = 1;
    this.players = [];

    this.titleBar = null;
    this.tableId = 0;
    this.shotLayer = null;
    
    this.tablePoker = [];
    this.tablePokerPic = {};
    
    this.lastShotPlayer = null;

    this.whoseTurn = 0;

};

PG.Game.prototype = {

    init: function(roomId) {
        this.roomId = roomId;
    },

    debug_log(obj) {
    console.log('*******');
    console.log(obj);
    console.log('********');
    },

	create: function () {
        this.stage.backgroundColor = '#1823456d3b'

        this.players.push(PG.createPlay(0, this));
        this.players.push(PG.createPlay(1, this));
        this.players.push(PG.createPlay(2, this));
        this.players[0].updateInfo(PG.playerInfo.uid, PG.playerInfo.username);
        PG.Socket.connect(this.onopen.bind(this), this.onmessage.bind(this), this.onerror.bind(this));

        this.createTitleBar();
	},
	
	onopen: function() {
	    console.log('socket onopen');
        PG.Socket.send([PG.Protocol.REQ_JOIN_ROOM, this.roomId]);
	},

    onerror: function() {
        console.log('socket connect onerror');
    },

	send_message: function(request) {
        PG.Socket.send(request);
	},
	
	onmessage: function(packet) {
	    let opcode = packet[0];
        let playerId = NaN;
	    switch(opcode) {
            case PG.Protocol.RSP_JOIN_ROOM:
                if (this.roomId === 1) {
                    PG.Socket.send([PG.Protocol.REQ_JOIN_TABLE, -1]);
                } else {
                    this.createTableLayer(packet[1]);
                }
                break;
            case PG.Protocol.RSP_TABLE_LIST:
                this.createTableLayer(packet[1]);
                break;
            case PG.Protocol.RSP_NEW_TABLE:
                this.tableId = packet[1];
                this.titleBar.text = '房间:' + this.tableId;
                break;
	        case PG.Protocol.RSP_JOIN_TABLE:
                this.tableId = packet[1];
                this.titleBar.text = '房间:' + this.tableId;
                let playerIds = packet[2];
                for (let i = 0; i < playerIds.length; i++) {
                    if (playerIds[i][0] === this.players[0].uid) {
                        let info_1 = playerIds[(i+1)%3];
                        let info_2 = playerIds[(i+2)%3];
                        this.players[1].updateInfo(info_1[0], info_1[1]);
                        this.players[2].updateInfo(info_2[0], info_2[1]);
                        break;
                    }
                }
                break;
            case PG.Protocol.RSP_DEAL_POKER:
                playerId=packet[1]
                let pokers = packet[2];
                console.log(pokers);
                this.dealPoker(pokers);
                this.whoseTurn = this.uidToSeat(playerId);
                this.startCallScore(0);
                break;
            case PG.Protocol.RSP_CALL_SCORE:
                playerId=packet[1]
                let score = packet[2];
                let callend = packet[3];
                this.debug_log(callend);
                this.whoseTurn = this.uidToSeat(playerId);
                //this.debug_log(playerId);

                let hanzi = ['不叫', "一分", "两分", "三分"];
                this.players[this.whoseTurn].say(hanzi[score]);
                if (!callend) {
                    this.whoseTurn = (this.whoseTurn + 1) % 3;
                    this.startCallScore(score);
                }
                break;
            case PG.Protocol.RSP_SHOW_POKER:
                this.whoseTurn = this.uidToSeat(packet[1]);
                this.tablePoker[0] = packet[2][0];
                this.tablePoker[1] = packet[2][1];
                this.tablePoker[2] = packet[2][2];
                this.players[this.whoseTurn].setLandlord();
                this.showLastThreePoker();
                break;
            case PG.Protocol.RSP_SHOT_POKER:
                this.handleShotPoker(packet);
                break;
            case PG.Protocol.RSP_GAME_OVER:
                let winner = packet[1];
                let coin = packet[2];

                let loserASeat = this.uidToSeat(packet[3][0]);
                this.players[loserASeat].replacePoker(packet[3], 1);
                this.players[loserASeat].reDealPoker();

                let loserBSeat = this.uidToSeat(packet[4][0]);
                this.players[loserBSeat].replacePoker(packet[4], 1);
                this.players[loserBSeat].reDealPoker();
//                 this.players[loserBSeat].removeAllPoker();
//               this.players[loserASeat].pokerInHand = [];

                this.whoseTurn = this.uidToSeat(winner);

                function gameOver() {
                    alert(this.players[this.whoseTurn].isLandlord ? "地主赢" : "农民赢");
                    PG.Socket.send([PG.Protocol.REQ_RESTART]);
                    this.cleanWorld();
                }
                this.game.time.events.add(3000, gameOver, this);
                break;
            case PG.Protocol.RSP_CHEAT:
                let seat = this.uidToSeat(packet[1]);
                this.players[seat].replacePoker(packet[2], 0);
                this.players[seat].reDealPoker();
                break;
            case PG.Protocol.RSP_RESTART:
                this.restart();
            default:
                console.log("UNKNOWN PACKET:", packet)
	    }
	},

    cleanWorld: function () {
        for (i =0; i < 3; i ++) {
            this.players[i].cleanPokers();
            try {
                this.players[i].uiLeftPoker.kill();
            }
            catch (err) {
            }
            this.players[i].uiHead.frameName = 'icon_farmer.png';
        }

        for (let i = 0; i < this.tablePoker.length; i++) {
                let p = this.tablePokerPic[this.tablePoker[i]];
                // p.kill();
                p.destroy();
            }
    },

	restart: function () {
        this.players = [];
        this.shotLayer = null;

        this.tablePoker = [];
        this.tablePokerPic = {};

        this.lastShotPlayer = null;

        this.whoseTurn = 0;

        this.stage.backgroundColor = '#182d3b';
        this.players.push(PG.createPlay(0, this));
        this.players.push(PG.createPlay(1, this));
        this.players.push(PG.createPlay(2, this));
        player_id = [1, 11, 12];
        for (let i = 0; i < 3; i++) {
            //this.players[i].uiHead.kill();
            this.players[i].updateInfo(player_id[i], ' ');
        }

        // this.send_message([PG.Protocol.REQ_DEAL_POKEER, -1]);
//        PG.Socket.send([PG.Protocol.REQ_JOIN_TABLE, this.tableId]);
	},

	update: function () {
	},

	uidToSeat: function (uid) {
	    for (let i = 0; i < 3; i++) {
//	        this.debug_log(this.players[i].uid);
	        if (uid === this.players[i].uid)
	            return i;
	    }
	    console.log('ERROR uidToSeat:' + uid);
	    return -1;
	},
    
    dealPoker: function(pokers) {

        for (let i = 0; i < 3; i++) {
            let p = new PG.Poker(this, 54, 54);
            this.game.world.add(p);
            this.tablePoker[i] = p.id;
            this.tablePoker[i + 3] = p;
        }

        for (let i = 0; i < 17; i++) {
            this.players[2].pokerInHand.push(54);
            this.players[1].pokerInHand.push(54);
            this.players[0].pokerInHand.push(pokers.pop());
        }

        this.players[0].dealPoker();
        this.players[1].dealPoker();
        this.players[2].dealPoker();
        //this.game.time.events.add(1000, function() {
        //    this.send_message([PG.Protocol.REQ_CHEAT, this.players[1].uid]);
        //    this.send_message([PG.Protocol.REQ_CHEAT, this.players[2].uid]);
        //}, this);
    },
     
    showLastThreePoker: function() {
        for (let i = 0; i < 3; i++) {
            let pokerId = this.tablePoker[i];
            let p = this.tablePoker[i + 3];
            p.id = pokerId;
            p.frame = pokerId;
            this.game.add.tween(p).to({ x: this.game.world.width/2 + (i - 1) * 60}, 600, Phaser.Easing.Default, true);
        }
        this.game.time.events.add(1500, this.dealLastThreePoker, this);
    },

    dealLastThreePoker: function() {
	    let turnPlayer = this.players[this.whoseTurn];

        for (let i = 0; i < 3; i++) {
            let pid = this.tablePoker[i];
            let poker = this.tablePoker[i + 3];
            turnPlayer.pokerInHand.push(pid);
            turnPlayer.pushAPoker(poker);
        }
        turnPlayer.sortPoker();
        if (this.whoseTurn === 0) {
            turnPlayer.arrangePoker();
            for (let i = 0; i < 3; i++) {
                let p = this.tablePoker[i + 3];
                let tween = this.game.add.tween(p).to({y: this.game.world.height - PG.PH * 0.8 }, 400, Phaser.Easing.Default, true);
                function adjust(p) {
                    this.game.add.tween(p).to({y: this.game.world.height - PG.PH /2}, 400, Phaser.Easing.Default, true, 400);
                };
                tween.onComplete.add(adjust, this, p);
            }
        } else {
            let first = turnPlayer.findAPoker(54);
            for (let i = 0; i < 3; i++) {
                let p = this.tablePoker[i + 3];
                p.frame = 54;
                p.frame = 54;
                this.game.add.tween(p).to({ x: first.x, y: first.y}, 200, Phaser.Easing.Default, true);
            }
        }

        this.tablePoker = [];
        this.lastShotPlayer = turnPlayer;
        if (this.whoseTurn === 0) {
            this.startPlay();
        }
    },

    handleShotPoker: function(packet) {
        this.whoseTurn = this.uidToSeat(packet[1]);
        let turnPlayer = this.players[this.whoseTurn];
        let pokers = packet[2];
        if (pokers.length === 0) {
            this.players[this.whoseTurn].say("不出");
        } else {
            let pokersPic = {};
            pokers.sort(PG.Poker.comparePoker);
            let count= pokers.length;
            let gap = Math.min((this.game.world.width - PG.PW * 2) / count, PG.PW * 0.36);
            for (let i = 0; i < count; i++) {
                let p = turnPlayer.findAPoker(pokers[i]);
                p.id = pokers[i];
                p.frame = pokers[i];
                p.bringToTop();
                this.game.add.tween(p).to({ x: this.game.world.width/2 + (i - count/2) * gap, y: this.game.world.height * 0.4}, 500, Phaser.Easing.Default, true);

                turnPlayer.removeAPoker(pokers[i]);
                pokersPic[p.id] = p;
            }
        
            for (let i = 0; i < this.tablePoker.length; i++) {
                let p = this.tablePokerPic[this.tablePoker[i]];
                // p.kill();
                p.destroy();
            }
            this.tablePoker = pokers;
            this.tablePokerPic = pokersPic;
            this.lastShotPlayer = turnPlayer;
            turnPlayer.arrangePoker();
        }
        if (turnPlayer.pokerInHand.length > 0) {
            this.whoseTurn = (this.whoseTurn + 1) % 3;
            if (this.whoseTurn === 0) {
                this.game.time.events.add(1000, this.startPlay, this);
            }
        }
    },

    startCallScore: function(minscore) {
        function btnTouch(btn) {
            this.send_message([PG.Protocol.REQ_CALL_SCORE, btn.score]);
            btn.parent.destroy();
            let audio = this.game.add.audio('f_score_' + btn.score);
            audio.play();
        };

        if (this.whoseTurn === 0) {
            let step = this.game.world.width/6;
            let ss = [1.5, 1, 0.5, 0];
            let sx = this.game.world.width/2 - step * ss[minscore];
            let sy = this.game.world.height * 0.6;
            let group = this.game.add.group();
            let pass = this.game.make.button(sx, sy, "btn", btnTouch, this, 'score_0.png', 'score_0.png', 'score_0.png');
            pass.anchor.set(0.5, 0);
            pass.score = 0;
            group.add(pass);
            sx += step;

            for (let i = minscore + 1; i <= 3; i++) {
                let tn = 'score_' + i + '.png';
                let call = this.game.make.button(sx, sy, "btn", btnTouch, this, tn, tn, tn);
                call.anchor.set(0.5, 0);
                call.score = i;
                group.add(call);
                sx += step;
            }
        } else {
            // TODO show clock on player
        }
        
    },

    startPlay: function() {
        if (this.isLastShotPlayer()) {
            this.players[0].playPoker([]);
        } else {
            this.players[0].playPoker(this.tablePoker);
        }
    },

    finishPlay: function(pokers) {
        this.send_message([PG.Protocol.REQ_SHOT_POKER, pokers]);
    },

    isLastShotPlayer: function() {
        return this.players[this.whoseTurn] === this.lastShotPlayer;
    },

    createTableLayer: function (tables) {
        tables.push([-1, 0]);

        let group = this.game.add.group();
        this.game.world.bringToTop(group);
        let gc = this.game.make.graphics(0, 0);
        gc.beginFill(0x00000080);
        gc.endFill();
        group.add(gc);
        let style = {font: "22px Arial", fill: "#fff", align: "center"};

        for (let i = 0; i < tables.length; i++) {
            let sx = this.game.world.width * (i%6 + 1)/(6 + 1);
            let sy = this.game.world.height * (Math.floor(i/6) + 1)/(4 + 1);

            let table = this.game.make.button(sx, sy, 'btn', this.onJoin, this, 'table.png', 'table.png', 'table.png');
            table.anchor.set(0.5, 1);
            table.tableId = tables[i][0];
            group.add(table);

            let text = this.game.make.text(sx, sy, '房间:' + tables[i][0] + '人数:' + tables[i][1], style);
            text.anchor.set(0.5, 0);
            group.add(text);

            if (i === tables.length - 1) {
                text.text = '新建房间';
            }
        }
    },

    quitGame: function () {
        this.state.start('MainMenu');
    },

    createTitleBar: function() {
        let style = {font: "22px Arial", fill: "#fff", align: "center"};
        this.titleBar = this.game.add.text(this.game.world.centerX, 0, '房间:', style);
    },

    onJoin: function (btn) {
        if (btn.tableId === -1) {
            this.send_message([PG.Protocol.REQ_NEW_TABLE]);
        } else {
            this.send_message([PG.Protocol.REQ_JOIN_TABLE, btn.tableId]);
        }
        btn.parent.destroy();
    }
};






