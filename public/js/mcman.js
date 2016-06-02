(function(window,document){
	B('#stop_link').on('click',function(){
		var oReq = new XMLHttpRequest();
		oReq.onload = function(){console.log(this.responseText);};
		oReq.open("post", "/api/v1/stop", true);
		oReq.send();
	});
	var menu=B('#menu'),
			WINDOW_CHANGE_EVENT=('onorientationchange' in window)?'orientationchange':'resize';
	function toggleHorizontal(){
		[].forEach.call(
				B('#menu .custom-can-transform'),
				function(el){
					el.classList.toggle('pure-menu-horizontal');
				}
		);
	};
	function toggleMenu(){
		if(menu[0].classList.contains('open')){
			setTimeout(toggleHorizontal,500);
		} else {
			toggleHorizontal();
		}
		menu.classList.toggle('open');
		B('#toggle').classList.toggle('x');
	};
	function closeMenu(){
		if(menu[0].classList.contains('open')){
			toggleMenu();
		}
	};
	B('#toggle').on('click', function(e){
		toggleMenu();
	});
	window.addEventListener(WINDOW_CHANGE_EVENT,closeMenu);
})(this,this.document);
