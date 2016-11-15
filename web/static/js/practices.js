$('.practice-info-body').hide();
$(function() {
	$('.action-show-practice-info').click(function() {
		console.log('Show '+ this.id);
		$('#'+ this.id +'.practice-info-toggle' ).toggle();
		$('#'+ this.id +'.practice-info-body').slideDown(500)
	});
	$('.action-hide-practice-info').click(function() {
		console.log('Hide ' + this.id);
		$('#'+ this.id +'.practice-info-toggle').toggle();
		$('#'+ this.id +'.practice-info-body').slideUp(500);
	});
});