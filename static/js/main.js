$(document).ready(function() {
  GetLastEvent();
});

var GetLastEvent = function() {
  const data = {action: "describe"};
  request(data, (res)=>{
    $("#last_event").text(res.last);
    $("#event_schedule").text(res.schedule);
  }, (e)=>{
    console.log(e.responseJSON.message);
  });
};

var Submit = function() {
  $("#submit").addClass('disabled');
  var minute = $('#minute').val();
  var hour = $('#hour').val();
  var day = $('#day').val();
  var month = $('#month').val();
  var year = $('#year').val();
  const data = {action: "put", minute, hour, day, month, year};
  request(data, (res)=>{
    $("#info").removeClass("hidden").addClass("visible");
    $("#submit").removeClass('disabled');
    window.setTimeout(() => location.reload(true), 1000);
  }, (e)=>{
    console.log(e.responseJSON.message);
    $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
    $("#submit").removeClass('disabled');
  });
};

var request = function(data, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};

var App = { url: location.origin + {{ .ApiPath }} };
