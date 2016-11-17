$('.practice-info-body').hide();
$(function() {
	$('.action-show-practice-info').click(function() {
		$('#' + this.id + '.remove-confirmation').hide(300);
		console.log('Show ' + this.id);
		$('#' + this.id + '.practice-info-toggle').toggle();
		$('#' + this.id + '.practice-info-body').slideDown(500)
	});
	$('.action-hide-practice-info').click(function() {
		$('#' + this.id + '.remove-confirmation').hide();
		console.log('Hide ' + this.id);
		$('#' + this.id + '.practice-info-toggle').toggle();
		$('#' + this.id + '.practice-info-body').slideUp(500);
	});
});

$(function() {
	$('.hide-content').click(function() {
		$('#' + this.id + '.remove-confirmation').hide(300);
		$target = this.id
		$.ajax({
			url : "hide/" + $target,
			success : function(result) {
				console.log('Hide ' + $target);
				$('#' + $target + '.visibility-toggle').toggle();
			}
		});
	});
	$('.publish-content').click(function() {
		$('#' + this.id + '.remove-confirmation').hide(300);
		$target = this.id
		$.ajax({
			url : "publish/" + $target,
			success : function(result) {
				console.log('Publish ' + $target);
				$('#' + $target + '.visibility-toggle').toggle();
			}
		});
	});
});

$(function() {
	$('.remove-content').click(function() {
		$('#' + this.id + '.remove-confirmation').toggle(300);
	});
});

$(function() {
	$('.remove-cancel-btn').click(function() {
		$('#' + this.id + '.remove-confirmation').toggle(300);
		console.log("TEST")
	});
	$(function() {
		$('.remove-confirmation-btn').click(function() {
			$target = this.id
			$.ajax({
				url : "remove/" + $target,
				success : function(result) {
					console.log('Remove ' + $target);
					$('#' + $target + '.practice-panel').remove();
				}
			});
		});
	});
});