<html>
    <head>
        <script>
            function changetext()
            {
                var newItemText=document.createTextNode("Choose file:")
                var newItemInput=document.createElement("input")
                newItemInput.name="uploadfile"
                newItemInput.type="file"

                var nametest=document.createTextNode("name:")
                var newNameInput=document.createElement("input")
                newNameInput.name="name"
                newNameInput.type="text"

                var addfile=document.getElementById("bt_addfile")
                var submit=document.getElementById("bt_submit")
                var newItemBr=document.createElement("br")

                var myform=document.getElementById("upform")
                myform.appendChild(newItemText);
                myform.appendChild(newItemInput);
                myform.appendChild(addfile);
                myform.appendChild(nametest);
                myform.appendChild(newItemBr);
                myform.appendChild(submit);
            }
        </script>
        <script type="text/javascript" src="/js/jquery.js" ></script>
    </head>
    <h1>Welcome to zimg world!</h1>
    <p>Upload image(s) to zimg:</p>
    <form enctype="multipart/form-data" action="upload" method=post target=_blank id="upform">
        Choose file:<input name="uploadfile" type="file">
        <input type="button" value="+" onclick="changetext()" id="bt_addfile">
        </br>
        相机名称:<input type="text" name="username">
        <input type="submit" value="upload" id="bt_submit">
    </form>

	<script type="application/javascript">
		//发送表单ajax请求
		$(':submit').on('click',function(){
			$.ajax({
				url:"buy",
				type:"POST",
				data:JSON.stringify($('form').serializeObject()),
				contentType:"application/json",  //缺失会出现URL编码，无法转成json对象
				success:function(){
					alert("成功");
				}
			});
		});

		/**
		 * 自动将form表单封装成json对象
		 */
		$.fn.serializeObject = function() {
			var o = {};
			var a = this.serializeArray();
			$.each(a, function() {
				if (o[this.name]) {
					if (!o[this.name].push) {
						o[this.name] = [ o[this.name] ];
					}
					o[this.name].push(this.value || '');
				} else {
					o[this.name] = this.value || '';
				}
			});
			return o;
		};
	</script>

    <p>More infomation: <a href="http://zimg.buaa.us">zimg.buaa.us</a></p>
</html>