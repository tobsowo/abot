var Profile = {
	signout: function(ev) {
		ev.preventDefault();
		cookie.removeItem("id");
		cookie.removeItem("session_token");
		m.route("/login");
	},
	data: function(uid) {
		return m.request({
			method: "GET",
			url: "/api/profile.json?uid=" + uid
		});
	},
	sendView: function(uid) {
		return m.request({
			method: "PUT",
			url: "/api/profile.json",
			data: { UserID: parseInt(uid, 10) }
		});
	}
};

Profile.controller = function() {
	if (cookie.getItem("id") === null) {
		return m.route("/login");
	}
	var _this = this;
	_this.username = m.prop("");
	_this.email = m.prop("");
	_this.phones = new List({type: Phone});
	_this.cards = new List({type: Card});
	var userId = cookie.getItem("id");
	Profile.data(userId).then(function(data) {
		_this.email(data.Email);
		_this.username(data.Name);
		_this.phones.userId(userId);
		_this.phones.data(data.Phones);
		_this.cards.userId(userId);
		_this.cards.data(data.Cards);
	}, function(err) {
		console.error(err);
	});
	// TODO implement for authentications
	// Profile.sendView(userId);
};

Profile.view = function(controller) {
	return m("div", {
		class: "body"
	}, [
		header.view(),
		Profile.viewFull(controller),
		Footer.view()
	]);
};

Profile.viewFull = function(controller) {
	return m("div", {
		id: "full",
		class: "container"
	}, [
		m("div", {
			class: "row margin-top-sm"
		}, [
			m("div", {
				class: "col-md-12"
			}, [
				m("h1", "Profile")
			])
		]),
		m("div", {
			class: "row"
		}, [
			m("div", {
				class: "col-md-7 margin-top-sm"
			}, [
				m("h3", "Account Details"),
				m("form", {
					class: "margin-top-sm"
				}, [
					m("div", {
						class: "card"
					}, [
						m("div", {
							class: "form-group"
						}, [
							m("label", "Username"),
							m("div", [
								m("div", controller.email())
							])
						]),
						m("div", {
							class: "form-group"
						}, [
							m("label", "Password"),
							m("div", [
								m("a", {
									href: "#"
								}, "Change password")
							])
						]),
						m("div", {
							class: "form-group"
						}, [
							m("label", {
								for: "username"
							}, "Name"),
							m("div", [
								m("input", {
									id: "username",
									type: "text",
									class: "form-control",
									value: controller.username()
								})
							])
						]),
						m("div", {
							class: "form-group margin-top-sm"
						}, [
							m("div", [
								m("a", {
									class: "btn btn-sm",
									href: "#/",
									onclick: Profile.signout
								}, "Sign out")
							])
						])
					]),
					m("h3", {
						class: "margin-top-sm"
					}, "Phone numbers"),
					m("div", {
						class: "form-group card"
					}, [
						m("div", [
							controller.phones.view()
						])
					]),
					m("h3", {
						class: "margin-top-sm"
					}, "Credit cards"),
					m("div", {
						class: "form-group card"
					}, [
						m("div", [
							controller.cards.view(),
							m("div", [
								m("a", {
									id: controller.cards.id + "-add-btn",
									class: "btn btn-sm",
									href: "/cards/new",
									config: m.route
								}, "+Add Card")
							])
						])
					])
				])
			])
		])
	]);
};
