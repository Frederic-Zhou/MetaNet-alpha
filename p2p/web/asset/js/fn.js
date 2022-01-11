//各页面记录page属性
var globalPage={};
if(window['layui']){
	layui.use(function(){
		window['form']=layui.form;
		window['layer']=layui.layer;
		window['laytpl']=layui.laytpl;
		window['laypage']=layui.laypage;
		window['layuidate']=layui.laydate;
		window['laytable']=layui.table;
	});
}
window['timezone_offset']=new Date().getTimezoneOffset()/60;
function runcb(fn,tm){
	if(typeof(fn)=='function'){
		if(!tm){tm=10;}
		setTimeout(function(){fn()},tm);
	}
}
function msgok(v,fn){layer.msg(v,{icon:1});runcb(fn,2000);}
function msgerr(v,fn){layer.msg(v,{icon:2,anim:6});runcb(fn,2000);}
function msgcry(v,fn){layer.msg(v,{icon:5,anim:6});runcb(fn,2000);}
function alertico(v,ico,fn){layer.alert(v,{icon:ico||'',anim:(ico==1)?0:6},function(idx){runcb(fn);layer.close(idx);});}
function alertok(v,fn){alertico(v,1,fn);}
function alerterr(v,fn){alertico(v,2,fn);}
function alertcry(v,fn){alertico(v,5,fn);}

function parseHash(){
	var hash=(location.hash).replace('#','');
	var arr=hash.split('?');
	if(arr.length<2){return {};}
	var obj={}
	var arr1=arr[1].split('&');
	$.each(arr1,function(k,v){
		var ss=v.split('=')
		obj[ss[0]]=ss[1];
	});
	return obj;
}

function trimform(el){
	return trimobj(form2obj(el));
}
function trimobj(obj){
	var dt={};
	$.each(obj,function(k,v){
		if(v==''){return true;}
		dt[k]=v;
	});
	return dt;
}
function form2obj_bak(el) {
	if (!el) {
		el = 'form:eq(0)';
	}
	var o = {};
	var a = $(el).serializeArray();
	$.each(a, function (i,v){
		var f=v.name,val=$.trim(v.value);
		if(o[f]!==undefined){
			if(!o[f].push) {
				o[f]=[o[f]];
			}
			o[f].push(val);
		} else {
			o[f]=val;
		}
	});
	return o;
}
function form2obj(el){
	if(!el){el='form:eq(0)';}
	var o={};
	var a= $(el).serializeArray();
	$.each(a,function(k,v){
		var f=v.name,vv=$.trim(v.value);
		if(f.indexOf('[]')>0){
			f=f.replace('[]','');
			if(o[f]===undefined){o[f]=[vv]}else{o[f].push(vv);}
		}else{
			if(o[f]!==undefined){
				if(!o[f].push){o[f]=[o[f]];}
				o[f].push(vv);
			}else{o[f]=vv;}
		}
	})
	return o;
}
function empty_obj(obj){
	var i=0;
	for(var k in obj){
		i++;break;
	}
	if(i>0){return false}
	return true;
}
//获取url参数
function get_param() {
	var url = location.search;
	var arr = url.replace(/^\?+/, "").split("&");
	var obj = {}
	for (var i = 0; i < arr.length; i++) {
		var v = arr[i].split('=')
		if (v.length > 1) {
			obj[v[0]] = decodeURIComponent(v[1])
		}
	}
	return obj;
}
function trip_slash(str){
	return str.replace(/\"/g,'')
}
function rndnum(n,m){
	if(m==undefined){return rndstr(n,'0123456789')}
	return (n+Math.round(Math.random()*(m-n)));
}
function rndstr(n,c){
	n=n||32;
	c=c||'ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678';	/*去掉了字符oOLl,9gq,Vv,Uu,I1*/
	var cl=c.length,v='';
	for(i=0;i<n;i++){
		v+=c.charAt(Math.floor(Math.random()*cl));
	}
	return v;
}
function orderno(){
	return layui.util.toDateString(new Date(),'yyyyMMddHHmmss')+rndstr(8,'0123456789')
}
function utf16to8(str) {
	var out, i, len, c;
	out = "";
	len = str.length;
	for (i = 0; i < len; i++) {
		c = str.charCodeAt(i);
		if ((c >= 0x0001) && (c <= 0x007F)) {
			out += str.charAt(i);
		} else if (c > 0x07FF) {
			out += String.fromCharCode(0xE0 | ((c >> 12) & 0x0F));
			out += String.fromCharCode(0x80 | ((c >> 6) & 0x3F));
			out += String.fromCharCode(0x80 | ((c >> 0) & 0x3F));
		} else {
			out += String.fromCharCode(0xC0 | ((c >> 6) & 0x1F));
			out += String.fromCharCode(0x80 | ((c >> 0) & 0x3F));
		}
	}
	return out;
}
function shownum(v){
	if($.isNumeric(v)) return v;
	return ''
}
// ================================================================
//密码复杂度
function chk_pwd(v){
	if(v==''){return;}
	if(!/^[\S]{8,30}$/.test(v)){return '密码长度不得小于8位';}
	if(!(/[0-9]/.test(v) && /[a-z]/.test(v) && /[A-Z]/.test(v))){return '必须包含数字和大小写字母';}
}
//IPV6
function isIPv6(str){
	return str.match(/:/g).length<=7 &&/::/.test(str)  ?/^([\da-f]{1,4}(:|::)){1,6}[\da-f]{1,4}$/i.test(str)  :/^([\da-f]{1,4}:){7}[\da-f]{1,4}$/i.test(str)
}
//ipv4
function chk_ipv4(str, item) {
	if (str == '') {return;}
	str = str.replace(/\n/g,',').replace(/\r/g,'');
	var arr = str.split(',');
	var pass = false;
	for (var i = 0; i < arr.length; i++) {
		var v = arr[i];
		v = v.replace(/\s/g, '')
		if(v.indexOf(':')>-1){//ipV6
			if(!isIPv6(v)){
				return '错误的IPv6格式';
			}
		}else{//ipV4
			if (/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(v)) {
				var pts = v.split('.');
				if (0 == (1 * pts[0])) {
					return '错误的IPv4格式';
				}
				if (pts[0] > 255 || pts[1] > 255 || pts[2] > 255 || pts[3] > 255) {
					return 'IP数值范围错误';
				}
			} else {
				return '错误的IPv4格式';
			}
		}
		
	}
}
//货币金额
function chk_currency(v, dom) {
	if (v == '') {
		return
	}
	if (1 * v == 0) {
		return
	}
	var pt = parseInt(window.moneyPoint)
	var regex = '^\\d+$';
	if (pt > 0) {
		regex = '^\\d+\\.?\\d{0,' + pt + '}$';
	}
	var re = new RegExp(regex);
	if (!re.test(v)) {
		if (/^[+-]?[\d]+([\.][\d]+)?([Ee][+-]?[\d]+)?$/.test(v)) {
			v = 1 * (new Number(v));
			dom.value = v;
			return chk_currency(v, dom)
		}
		return 'Wrong Currency format!';
	}
}
//千分比
function chk_rate(v, item) {
	if (v == '') {return}
	v = 1 * v;
	if (1 * v == 0) {
		return
	}
	if (v < 0 || v > 1000) {
		return '比例范围是1~1000'
	}
}
// ================================================================
function add_zero(v, len) {
	var str = '0000000000' + v;
	var alen = str.length;
	return str.substr(alen - len, len);
}
//格式化金额
function fix_money(v){
	if (('' + v).indexOf('.') < 0) {
		return ('' + v).replace(/\d(?=(?:\d{3})+\b)/g, '$&,');
	}
	var arr=(''+v).split('.');
	return ('' + arr[0]).replace(/\d(?=(?:\d{3})+\b)/g, '$&,') + '.'+arr[1];
}
//格式化日期
function render_date(d){
	if(!d){return ''}
	var dt=new Date(d);
	return layui.util.toDateString(dt,'MM-dd HH:mm:ss')
	return (d.substr(0,19)).replace(/T/g,' ').replace(/\+.+/g,'').replace(/Z/,'');
}
//格式化是否
function render_bool(v, yes, no) {
	return (true==v)?'<font color="green">'+(!!yes ? yes :'是 √') + '</font>':'<font color="red">'+(!!no?no:'否 ×')+'</font>';
}
// 该方法用于解决,使用fixed固定列后,行高和其他列不一致的问题
function fix_tr_height(){
	$(".layui-table-main tr").each(function (index ,val) {
		$($(".layui-table-fixed .layui-table-body tbody tr")[index]).height($(val).height());
	});
	$(".layui-table-header tr").each(function (index ,val) {
		$($(".layui-table-fixed .layui-table-header thead tr")[index]).height($(val).height());
	});
}

function render_bgc(){
	$('[bgc]').each(function(){
		var zs=$(this);
		var bgc=zs.attr('bgc');
		zs.parent().parent().css('background-color',bgc);
	})
}
// ================================================================
function dialog(title,content,fn){
	var dlg_idx=layer.open({
		type:1,anim :-1,isOutAnim:false
		,title:title||'标题'
		,content:content||''
		,area:['80%','80%']
		,btn:'关闭'
		,success:function(layero,index){
			if(typeof(fn)=='function'){fn()}
		}
	})
}
function load_frame(url,title,fn){
	var dlg_idx=layer.open({
		type:2,anim :-1,isOutAnim:false
		,maxmin:true
		,title:title||'标题'
		,content:url||''
		,area:['80%','88%']
		,btn:'关闭'
		,success:function(layero,index){
			if(typeof(fn)=='function'){fn()}
		}
	})
}
function loading(){
	return layer.load(1,{time:10*1000});
}
function closeAll(){
	layer.closeAll();
}
function closeUp(){
	try{
		window.parent.reload_table();
		window.parent.layer.closeAll();
	}catch(e){
		layer.closeAll();
	}
}
function layconfirm(str,fn1,fn2){
	layer.confirm(str,{icon:3,title:'询问'},function(idx){
		layer.close(idx);
		if(typeof(fn1)=='function'){fn1();}
	},function(idx){
		layer.close(idx);
		if(typeof(fn2)=='function'){fn2();}
	});
}
function laydate(elem,type,fmt,opts){
	var opt={elem:elem}
	opt.type=type||'datetime';
	opt.format=fmt||'yyyy-MM-dd HH:mm:ss';
	if(opts){
		for(var k in opts){opt[k]=opts[k];}
	}
	layuidate.render(opt)
}
// ================================================================
function ajax_form(action,dom,fn){
	var dt=form2obj1(dom);
	var idx=loading();
	var data={action:action,data:JSON.stringify(dt)};
	$.ajax({
		type: "POST",
		url:'/admin/action',
		dataType: "json",
		data: JSON.stringify(data),
		processData: false,
		contentType: 'application/json',
	}).done(function(res){
		layer.close(idx);
		if(typeof(fn)=='function'){
			fn(res);return;
		}
		if(res.Message=='success'){
			alertok('数据处理成功',function(){
				closeAll();
				reload_table();
			});
		}else{
			alerterr('错误信息：<br/>'+res.Data);
		}
	}).fail(function(){
		layer.close(idx);
		alertcry('网络请求出错');
	})
}

// ================================================================