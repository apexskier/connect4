$(function() {
    var user = null;
    function getUser() {
        return $.get('/api/user').then(function(data) {
            if (data) {
                data = JSON.parse(data);
                user = data.user;
                if (data.game >= 0) {
                    currentGameService.setId(data.game);
                }
            } else {
                user = null;
            }
        }).fail(function(err) {
            user = null;
        });
    }
    function apiFail(err) {
        console.log(err);
        alert(err.responseText.trim() || err.statusText);
        if (err.status == 404 || err.status == 500) {
            appRouter.navigate('index', {trigger: true});
        }
    }
    var AppRouter = Backbone.Router.extend({
        routes: {
            'login':        'login',
            'games':        'games',
            'play/:gameid': 'play',
            '':             'index',
            '*anything':    'index'
        },
        index: function() {
            if (this.validate_loggedin()) {
                if (currentGameService.get().hasOwnProperty('id') && currentGameService.get().id) {
                    this.navigate('play/' + currentGameService.get().id, {trigger: true});
                } else {
                    this.navigate('games', {trigger: true});
                }
            } else {
                this.navigate('login', {trigger: true});
            }
        },
        login: function() {
            this.swapView(new LoginView());
        },
        games: function() {
            if (this.validate_loggedin()) {
                this.swapView(new GamesView());
            } else {
                this.navigate('login', {trigger: true});
            }
        },
        play: function(gameid) {
            var that = this;
            if (this.validate_loggedin() && gameid) {
                $.ajax({
                    type: 'GET',
                    url: '/api/games/' + gameid,
                    contentType: 'application/json'
                }).done(function(data) {
                    data = JSON.parse(data);
                    data.board = false;
                    if (!data.hasOwnProperty(data.player2)) {
                        data.player2 = false;
                    }
                    that.swapView(new PlayView({model: data}));
                }).fail(apiFail);
            } else {
                this.navigate('login', {trigger: true});
            }
        },

        validate_loggedin: function() {
            return user;
        },

        swapView: function(view) {
            if (this.currentView) {
                this.currentView.remove();
            }
            this.currentView = view;

            $('#entry').html(this.currentView.render().el);
        }
    });

    var AppView = Backbone.View.extend({
        el: $('body'),
        events: {
            'click #logout': 'logout',
            'click #games': 'games',
            'click #new-game': 'new_game'
        },
        render: function() {
            this.$('#content').text('Loading...');
            return this;
        },
        initialize: function() {
            this.listenTo(currentGameService.get(), 'start', function(id) {
                appRouter.navigate('play/' + id, {trigger: true});
            });
            this.listenTo(currentGameService.get(), 'change', function(id) {
                if (typeof id !== 'number') id = id.id;
                appRouter.navigate('play/' + id, {trigger: true});
            });
        },
        logout: function() {
            $.ajax({
                type: 'DELETE',
                url: '/api/user',
                contentType: 'application/json'
            }).always(function() {
                user = null;
                userView.render();
                eraseCookie('connect4id');
                appRouter.navigate('index', {trigger: true});
            });
            return this;
        },
        games: function () {
            appRouter.navigate('games', {trigger: true});
            return this;
        },
        new_game: function () {
            $.ajax({
                type: 'POST',
                url: '/api/games',
                contentType: 'application/json'
            }).done(function(data) {
                data = JSON.parse(data);
                console.log(data);
                appRouter.navigate('play/' + data.id, {trigger: true});
            }).fail(apiFail);
            return this;
        },
    });

    var UserView = Backbone.View.extend({
        el: $('#user'),
        template: _.template($('#user-view').html(), {variable: 'user'}),
        render: function() {
            this.$el.html(this.template(user));
            return this;
        }
    });
    var GamesView = Backbone.View.extend({
        events: {
            'click .play-game': 'play_game'
        },
        template: _.template($('#games-view').html()),
        initialize: function() {
            this.listenTo(eventPendingGames, 'add', this.add_game);
        },
        render: function() {
            this.$el.html(this.template({games: pendingGames})).append(new EmptyGamesView().render().el);
            return this;
        },
        add_game: function(ev) {
            this.$('#games-available').append(new AvailableGameView({model: ev}).render().el);
        },
        play_game: function(e) {
            var $target = $(e.target).closest('.play-game');
            var gameid = $target.data('gameid');
            $.ajax({
                type: 'POST',
                url: '/api/games/' + gameid,
                contentType: 'application/json'
            }).done(function(data) {
                data = JSON.parse(data);
                appRouter.navigate('play/' + gameid, {trigger: true});
            }).fail(apiFail);
        }
    });
    var EmptyGamesView = Backbone.View.extend({
        template: _.template($('#empty-games-view').html()),
        initialize: function() {
            this.listenTo(eventPendingGames, 'add', this.render);
            this.listenTo(eventPendingGames, 'remove', this.render);
        },
        render: function() {
            this.$el.html(this.template({l: Object.keys(pendingGames).length}));
            return this;
        }
    });
    var ListItemView = Backbone.View.extend({
        tagName: 'li',
        template: _.template($('#list-item-template').html()),
        initialize: function() {
            this.listenTo(this.model, 'remove', this.remove);
        },
        render: function() {
            this.$el.html(this.template(this.model));
            return this;
        }
    });
    var AvailableGameView = ListItemView.extend({
        template: _.template($('#available-game-view').html()),
    });
    var LoginView = Backbone.View.extend({
        template: _.template($('#login-view').html()),
        events: {
            'submit #login-form': 'login'
        },
        render: function () {
            this.$el.html(this.template());
            return this;
        },
        login: function(e) {
            e.preventDefault();
            var username = this.$('#login-username').val();
            eraseCookie('connect4id');
            $.ajax({
                type: 'POST',
                url: '/api/users',
                contentType: 'application/json',
                data: JSON.stringify(username)
            }).done(function(data) {
                data = JSON.parse(data);
                user = data;
                console.log(data);
                getUser().then(function() {
                    appRouter.navigate('index', {trigger: true});
                    userView.render();
                });
            }).fail(apiFail);
        }
    });
    var PlayView = Backbone.View.extend({
        template: _.template($('#play-view').html()),
        initialize: function() {
            currentGameService.setId(this.model.id);
            this.listenTo(currentGameService.get(), 'change', this.render);
            this.listenTo(currentGameService.get(), 'start', this.render);
            this.listenTo(currentGameService.get(), 'stop', this.leave);
        },
        events: {
            'click .col': 'move',
            'touch .col': 'move',
            'click #leave-game': 'leave_click'
        },
        leave: function(e) {
            $.ajax({
                type: 'DELETE',
                url: '/api/games/' + this.model.id,
                contentType: 'application/json'
            }).always(function(data) {
                appRouter.navigate('index', {trigger: true});
            });
        },
        leave_click: function(e) {
            currentGameService.setId(null);
        },
        render: function(model) {
            if (model) {
                this.model = model;
            }
            this.$el.html(this.template({game: this.model, user: user}));
            return this;
        },
        move: function(e) {
            var $target = $(e.target).closest('.col');
            var col = $target.data('col');
            $.ajax({
                type: 'PUT',
                url: '/api/games/' + this.model.id,
                contentType: 'application/json',
                data: JSON.stringify({
                    gameid: this.model.id,
                    row: col
                })
            }).done(function(data) {
                data = JSON.parse(data);
                console.log(data);
                if (data.status == "illegal") {
                    alert(data.details);
                } else if (data.status == "success") {
                    currentGameService.merge(data.game);
                } else if (data.status == "win") {
                    currentGameService.merge(data.game);
                    alert("you won!")
                    appRouter.navigate('games', {trigger: true});
                }
            }).fail(apiFail);
        }
    });

    var pendingGames = {};
    var eventPendingGames = {};
    _.extend(eventPendingGames, Backbone.Events);
    function getPendingGames() {
        $.get('/api/games').done(function(data) {
            var nextGames = _.filter(JSON.parse(data), function(game) {
                if (!user) return false;
                return game.player1 != user.username;
            });
            var nextIds = [];
            _.each(nextGames, function(game) {
                if (!_.has(pendingGames, game.id)) {
                    _.extend(game, Backbone.Events);
                    pendingGames[game.id] = game;
                    eventPendingGames[game.id] = game;
                    eventPendingGames.trigger('add', game);
                }
                nextIds.push(game.id);
            });
            pendingGames = _.omit(pendingGames, function(game, id) {
                var val = _.contains(nextIds, id);
                if (val) {
                    eventPendingGames.trigger('remove');
                    game.trigger('remove');
                    delete eventPendingGames[id];
                }
                return val;
            });
        }).fail(function(err) {
            pendingGames = {};
            eventPendingGames = {};
            _.extend(eventPendingGames, Backbone.Events);
            console.log(err);
        });
    }
    var gamesPoll = setInterval(getPendingGames, 1000);
    getPendingGames();

    function GameService() {
        var game = {};
        _.extend(game, Backbone.Events);
        var currentId = null;
        var poller = null;
        return {
            get: get,
            setId: set,
            merge: merge
        }

        function get() {
            return game;
        }

        function set(id) {
            game.id = null;
            if (id === null) {
                game.trigger('stop');
                clearInterval(poller);
            } else {
                game.trigger('change', id)
            }
            if (currentId === null) {
                game.trigger('start', id);
                poller = setInterval(poll, 1000);
                poll();
            }
            currentId = id;
        }

        function merge(data) {
            if (!data.hasOwnProperty('player2')) data.player2 = false;
            _.extend(game, data);
            for (var col = 0; col < game.board.cols; col++) {
                if (!game.board.slots.hasOwnProperty(col)) game.board.slots[col] = []
                for (i = game.board.slots[col].length; i < game.board.rows; i++) {
                    game.board.slots[col].push(false);
                }
                game.board.slots[col].reverse()
            }
            game.trigger('change', data)
        }

        function poll() {
            if (currentId) {
                $.get('/api/games/' + currentId).done(function(data) {
                    data = JSON.parse(data);
                    merge(data);
                }).fail(function(err) {
                    console.log(err);
                    alert("You lost.");
                    set(null);
                    appRouter.navigate('index', {trigger: true});
                    if (err.status == 0) {
                        user = null;
                    }
                });
            }
        }
    }

    var currentGameService = new GameService();
    var app = new AppView();
    var userView;
    var appRouter = new AppRouter();
    getUser().then(function() {
        userView = new UserView();
        userView.render();
        Backbone.history.start();
    });
});
