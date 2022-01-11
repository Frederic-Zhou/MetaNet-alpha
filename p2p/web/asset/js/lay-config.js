window.rootPath = (function (src) {
	src = document.scripts[document.scripts.length - 1].src;
	return src.substring(0, src.lastIndexOf("/") + 1);
})();

layui.config({
	base: rootPath + "lay-module/",
	version: false
}).extend({
	miniMenu: "layuimini/miniMenu", // layuimini菜单扩展
	miniTab: "layuimini/miniTab", // layuimini tab扩展
	miniTheme: "layuimini/miniTheme", // layuimini 主题扩展
	miniAdmin: "layuimini/miniAdmin", // layuimini后台扩展
});